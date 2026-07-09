package daemon

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
	"github.com/NordSecurity/nordvpn-linux/snapconf"
	"github.com/godbus/dbus/v5"
	"github.com/snapcore/snapd/client"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/encoding/prototext"
)

const (
	// maxDaemonLogSize is the maximum size of daemon logs to collect (500 MB).
	maxDaemonLogSize = 500 * 1024 * 1024

	// maxZipFileSize caps the resulting diagnostics archive (40 MB).
	maxZipFileSize = 40 * 1024 * 1024
	logPrefix      = "[troubleshoot]"

	// User-facing messages sent via pb.DiagnosticsProgress.Error. Centralised
	// here so support can grep them, and so we never accidentally diverge two
	// copies of the same wording.
	zipTooLargeMsg       = "Diagnostics file exceeds 40 MB limit. Please contact support for assistance."
	failedCreateZipMsg   = "Failed to create zip file: %v"
	failedChownZipMsg    = "Failed to change file ownership: %v"
	failedCollectMsg     = "Failed to collect diagnostics: %v"
	failedCloseZipMsg    = "Failed to close zip file: %v"
	noDaemonLogSourceMsg = "We couldn't extract daemon logs automatically because the daemon was not started via systemd or snap. Contact our support team for help collecting logs manually."
)

// errZipSizeLimitExceeded is returned when writing more data would push the
// diagnostics zip past maxZipFileSize.
var errZipSizeLimitExceeded = errors.New("diagnostics zip exceeds size limit")

func (r *RPC) CollectDiagnostics(in *pb.Empty, srv pb.Daemon_CollectDiagnosticsServer) error {
	caller, err := resolveDiagnosticsCaller(srv.Context())
	if err != nil {
		log.Error(logPrefix, "troubleshot failed with:", err)
		return srv.Send(&pb.DiagnosticsProgress{Error: err.Error()})
	}

	// createDiagnosticsZip uses os.CreateTemp under the hood, so we always
	// get a fresh file with no chance of collision — no existence check, no
	// TOCTOU race, no overwrite of an old report, even if two collections
	// land in the same second.
	zipFile, err := createDiagnosticsZip(caller.outputDir)
	if err != nil {
		log.Error(logPrefix, "failed to create diagnostics zip:", err)
		return srv.Send(&pb.DiagnosticsProgress{Error: fmt.Sprintf(failedCreateZipMsg, err)})
	}
	zipPath := zipFile.Name()

	if snapconf.IsUnderSnap() {
		if err := os.Chmod(zipPath, internal.PermUserRWGroupROthersR); err != nil {
			log.Error(logPrefix, "failed to change file permissions:", err)
		}
	} else {
		if err := os.Chown(zipPath, int(caller.uid), int(caller.gid)); err != nil {
			log.Error(logPrefix, "failed to change file ownership:", err)
			return abortDiagnosticsWithMsg(zipFile, srv, fmt.Sprintf(failedChownZipMsg, err))
		}
	}

	state := r.collectAppState(srv.Context())
	if err := collectDiagnosticsData(srv, zipFile, caller.user.HomeDir, state); err != nil {
		if errors.Is(err, errZipSizeLimitExceeded) {
			log.Error(logPrefix, "diagnostics zip exceeded 40 MB limit")
			return abortDiagnosticsWithMsg(zipFile, srv, zipTooLargeMsg)
		}
		log.Error(logPrefix, "failed to collect diagnostics:", err)
		return abortDiagnosticsWithMsg(zipFile, srv, fmt.Sprintf(failedCollectMsg, err))
	}

	if err := zipFile.Close(); err != nil {
		log.Error(logPrefix, "failed to close zip file:", err)
		if removeErr := os.Remove(zipPath); removeErr != nil {
			log.Error(logPrefix, "failed to delete zip", removeErr)
		}
		return srv.Send(&pb.DiagnosticsProgress{Error: fmt.Sprintf(failedCloseZipMsg, err)})
	}

	return srv.Send(&pb.DiagnosticsProgress{
		Step:     "Done",
		FilePath: zipPath,
	})
}

// abortDiagnosticsWithMsg closes and deletes the partial zip file, then sends
// a user-facing error on the stream.
func abortDiagnosticsWithMsg(zipFile *os.File, srv pb.Daemon_CollectDiagnosticsServer, msg string) error {
	_ = zipFile.Close()
	if err := os.Remove(zipFile.Name()); err != nil {
		log.Error(logPrefix, "failed to delete zip", err)
	}
	return srv.Send(&pb.DiagnosticsProgress{Error: msg})
}

// createDiagnosticsZip atomically creates a uniquely-named diagnostics zip
// inside outputDir. The filename embeds a second-precision timestamp plus a
// random suffix from os.CreateTemp's `*` substitution, guaranteeing
// different paths even for back-to-back calls within the same second.
func createDiagnosticsZip(outputDir string) (*os.File, error) {
	pattern := fmt.Sprintf("nordvpn-diagnostics-%s-*.zip", time.Now().Format("20060102-150405"))
	return os.CreateTemp(outputDir, pattern)
}

// appState bundles the daemon's view of itself for inclusion in system-info.txt.
// Captured up front so addSystemInfo doesn't depend on the *RPC.
type appState struct {
	version  string
	status   string
	settings string
}

// collectAppState pulls the daemon's version, status, and settings via the
// existing in-process RPC handlers, formatted as multi-line text blocks
// (prototext.Format). Errors are rendered inline so the corresponding block
// is never silently empty.
func (r *RPC) collectAppState(ctx context.Context) appState {
	out := appState{version: r.version + "\n"}

	if status, err := r.Status(ctx, &pb.Empty{}); err != nil {
		out.status = fmt.Sprintf("status error: %v\n", err)
	} else {
		out.status = prototext.Format(status)
	}

	if settings, err := r.Settings(ctx, &pb.Empty{}); err != nil {
		out.settings = fmt.Sprintf("settings error: %v\n", err)
	} else {
		out.settings = prototext.Format(settings)
	}
	return out
}

// daemonSupervisor identifies how the nordvpn daemon is being managed on the
// host, so addDaemonLogs can pick the matching log source. Detection runs
// once at collection time (detectDaemonSupervisor); addDaemonLogs itself is a
// pure dispatch on this value, which keeps it trivially testable.
type daemonSupervisor int

const (
	daemonSupervisorUnknown daemonSupervisor = iota
	daemonSupervisorSnap
	daemonSupervisorSystemd
)

// diagnosticsCaller bundles the identity of the client that invoked
// CollectDiagnostics (resolved from the gRPC peer credentials) together with
// the directory where the diagnostics zip should be written.
type diagnosticsCaller struct {
	user      *user.User
	uid       uint32
	gid       uint32
	outputDir string
}

// resolveDiagnosticsCaller extracts the caller's UID/GID from the gRPC
// context, looks up their user record, and picks the directory where the
// diagnostics zip will land. The actual filename is generated atomically by
// os.CreateTemp at write time.
func resolveDiagnosticsCaller(ctx context.Context) (*diagnosticsCaller, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to get peer from context")
	}
	cred, ok := p.AuthInfo.(internal.UcredAuth)
	if !ok {
		return nil, fmt.Errorf("failed to get credentials from peer")
	}
	userInfo, err := user.LookupId(strconv.FormatUint(uint64(cred.Uid), 10))
	if err != nil {
		return nil, fmt.Errorf("failed to lookup user: %w", err)
	}
	return &diagnosticsCaller{
		user:      userInfo,
		uid:       cred.Uid,
		gid:       cred.Gid,
		outputDir: resolveOutputDir(userInfo.HomeDir),
	}, nil
}

// resolveOutputDir picks the directory for the diagnostics zip: the user's
// Downloads folder when present, falling back to their home directory. If
// the chosen directory is a symlink, /tmp is used instead to avoid writing
// through user-controlled symlinks.
func resolveOutputDir(homeDir string) string {
	if snapconf.IsUnderSnap() {
		return snapconf.LogsDir()
	}
	outputDir := homeDir
	downloadsDir := filepath.Join(homeDir, "Downloads")
	if internal.FileWritable(downloadsDir) && !internal.IsSymLink(downloadsDir) {
		outputDir = downloadsDir
	}
	if !internal.FileWritable(outputDir) || internal.IsSymLink(outputDir) {
		outputDir = "/tmp"
	}

	return outputDir
}

// detectDaemonSupervisor probes the host to figure out which supervisor is
// running the daemon. Order matters: snap takes precedence over systemd
// because snap-confined builds also see /run/systemd/system.
func detectDaemonSupervisor() daemonSupervisor {
	switch {
	case snapconf.IsUnderSnap():
		return daemonSupervisorSnap
	case internal.IsSystemd():
		return daemonSupervisorSystemd
	default:
		return daemonSupervisorUnknown
	}
}

// sizeLimitedWriter wraps an io.Writer and caps total bytes written to limit.
// Writes that would push past the cap are truncated to the remaining space:
// the prefix is forwarded to the underlying writer and errZipSizeLimitExceeded
// is returned alongside the partial count, so callers see exactly how much
// was accepted.
type sizeLimitedWriter struct {
	w       io.Writer
	limit   int64
	written int64
}

func (lw *sizeLimitedWriter) Write(p []byte) (int, error) {
	// Subtraction avoids int64 overflow that a naive `written+len(p) > limit`
	// could hit when written is close to math.MaxInt64.
	remaining := lw.limit - lw.written
	if remaining <= 0 {
		return 0, errZipSizeLimitExceeded
	}
	if int64(len(p)) <= remaining {
		n, err := lw.w.Write(p)
		lw.written += int64(n)
		return n, err
	}
	n, err := lw.w.Write(p[:remaining])
	lw.written += int64(n)
	if err != nil {
		return n, err
	}
	return n, errZipSizeLimitExceeded
}

// diagnosticsStep represents one unit of diagnostics collection: the message
// shown to the user and the function that performs the work.
//
// When fatal is true, any error returned by collect aborts the whole RPC; the
// partial zip is then discarded by the caller. Non-fatal step errors are
// logged as warnings so the remaining sections still make it into the report.
type diagnosticsStep struct {
	description string
	collect     func() error
	fatal       bool
}

// collectDiagnosticsData creates the zip writer and runs each collection
// step, reporting progress to the client as "[n/total] description".
// Individual step failures are logged as warnings and do not abort collection
// — partial diagnostics are still useful. Size-limit overflow is fatal and is
// returned to the caller so the partial zip can be discarded.
func collectDiagnosticsData(
	srv pb.Daemon_CollectDiagnosticsServer,
	output io.Writer,
	homeDir string,
	state appState,
) error {
	limited := &sizeLimitedWriter{w: output, limit: maxZipFileSize}
	zipWriter := zip.NewWriter(limited)
	defer zipWriter.Close() // nolint:errcheck

	// Buffered so the log_extraction_report.log entry can be written as the last zip
	// entry; zip.Writer finalizes each entry when the next Create is called,
	// so we can't keep this entry open while other entries are being written.
	var logExtractionReport bytes.Buffer
	logf := func(format string, args ...any) {
		fmt.Fprintf(&logExtractionReport, "%s %s\n",
			time.Now().Format("2006/01/02 15:04:05"),
			fmt.Sprintf(format, args...))
	}
	logf("diagnostics collection started (version=%s)", strings.TrimSpace(state.version))

	steps := []diagnosticsStep{
		{"Collecting daemon logs...", func() error {
			return addDaemonLogs(zipWriter, detectDaemonSupervisor())
		}, true},
		{"Collecting CLI logs...", func() error {
			if snapconf.IsUnderSnap() {
				// NOTE: AppArmor's home interface uses owner-qualified rules: nordvpnd
				// (root) cannot read paths owned by the user's UID, so these
				// files are inaccessible from daemon context under snap.
				// We are skipping it for now.
				return nil
			}
			cliLog := filepath.Join(homeDir, ".config", "nordvpn", "cli.log")
			return addFileToZip(zipWriter, cliLog, "cli.log")
		}, false},
		{"Collecting user logs...", func() error {
			if snapconf.IsUnderSnap() {
				// NOTE: AppArmor's home interface uses owner-qualified rules: nordvpnd
				// (root) cannot read paths owned by the user's UID, so these
				// files are inaccessible from daemon context under snap.
				// We are skipping it for now.
				return nil
			}
			cacheDir := filepath.Join(homeDir, ".cache", "nordvpn")
			return addDirectoryToZip(zipWriter, cacheDir, "cache")
		}, false},
		{"Collecting system info...", func() error {
			return addSystemInfo(zipWriter, state)
		}, false},
		{"Collecting network info...", func() error {
			return addNetworkInfo(zipWriter)
		}, false},
		{"Collecting DNS info...", func() error {
			return addDNSInfo(zipWriter)
		}, false},
		{"Collecting firewall rules...", func() error {
			return addNFTablesInfo(zipWriter)
		}, false},
	}

	total := len(steps)
	for i, step := range steps {
		desc := fmt.Sprintf("[%d/%d] %s", i+1, total, step.description)
		if err := srv.Send(&pb.DiagnosticsProgress{Step: desc}); err != nil {
			log.Warn(logPrefix, "failed to report the progress", err)
		}
		logf("step started: %s", step.description)
		err := step.collect()
		if err == nil {
			logf("step completed: %s", step.description)
			continue
		}
		if step.fatal {
			logf("step failed (fatal): %s: %v", step.description, err)
			writeLogExtractionReport(zipWriter, &logExtractionReport)
			return err
		}
		if errors.Is(err, errZipSizeLimitExceeded) {
			logf("step failed (size limit exceeded): %s: %v", step.description, err)
			writeLogExtractionReport(zipWriter, &logExtractionReport)
			return err
		}
		logf("step failed: %s: %v", step.description, err)
		log.Info(logPrefix, "diagnostics step failed:", step.description, err)
	}

	logf("diagnostics collection finished")
	writeLogExtractionReport(zipWriter, &logExtractionReport)

	// Explicit finalize so caller sees any central-directory write error
	// (e.g. hitting the size cap on the last flush).
	return zipWriter.Close()
}

// writeLogExtractionReport flushes the in-memory progress log into the zip as
// log_extraction_report.log. Best-effort: any failure here is logged to stderr and
// swallowed, since we don't want a log-writing error to mask the original
// collection outcome.
func writeLogExtractionReport(zipWriter *zip.Writer, buf *bytes.Buffer) {
	w, err := zipWriter.Create("log_extraction_report.log")
	if err != nil {
		log.Info(logPrefix, "failed to create log_extraction_report.log:", err)
		return
	}
	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Info(logPrefix, "failed to write log_extraction_report.log:", err)
	}
}

// stdoutAsRegularFile returns the path /proc/self/fd/1 resolves to when the
// daemon's stdout is a regular file, and false otherwise (tty, pipe, socket,
// symlink, or unreadable). Lstat (not Stat) is used so symlinks are rejected
// up front — we don't want to read through a user-controlled link. Used as a
// last-resort log source when no supervisor was detected.
func stdoutAsRegularFile() (string, bool) {
	target, err := os.Readlink("/proc/self/fd/1")
	if err != nil {
		return "", false
	}
	info, err := os.Lstat(target)
	if err != nil {
		return "", false
	}
	if !info.Mode().IsRegular() {
		return "", false
	}
	return target, true
}

// addDaemonLogs writes the daemon's logs to daemon.log inside the archive,
// dispatching on supervisor to pick the right log source. The zip writer
// applies its own deflate compression, so no extra gzip layer is needed.
// The unknown variant is fatal: nothing is written and an explanatory error
// is returned to the caller for surfacing to the user.
func addDaemonLogs(zipWriter *zip.Writer, supervisor daemonSupervisor) error {
	writer, err := zipWriter.Create("daemon.log")
	if err != nil {
		return err
	}

	// Two separate paths because inside snap confinement neither `journalctl`
	// nor sdjournal can reach the host journal — they only see the snap's
	// private mount-namespace view. snapd's own Logs API is the only thing on
	// the snap side with privileged host-journal access. Outside snap we
	// shell out to `journalctl -r` (newest first) so the 500 MB cap drops the
	// *oldest* entries instead of the newest.
	switch supervisor {
	case daemonSupervisorSnap:
		return streamSnapLogs(writer, "nordvpn.nordvpnd")
	case daemonSupervisorSystemd:
		return streamCommandToWriter(writer, "journalctl", "-u", "nordvpnd", "--no-pager", "-r")

	case daemonSupervisorUnknown:
		fallthrough
	default:
		// Last-resort fallback: if the daemon's own stdout (fd 1) is a
		// regular file, the operator likely redirected logs there manually
		// (e.g. `nordvpnd > /var/log/nordvpnd.log`, custom unit, container
		// with stdout pinned to a file). Streaming that file gives support
		// something to work with even when no supervisor was detected.
		if path, ok := stdoutAsRegularFile(); ok {
			return streamFileToWriter(writer, path)
		}
		return errors.New(noDaemonLogSourceMsg)
	}
}

// streamCommandToWriter runs a command and streams its stdout to writer, capped
// at maxDaemonLogSize bytes. The process is always killed and reaped on return
// so the caller doesn't have to deal with orphaned children.
func streamCommandToWriter(writer io.Writer, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	n, err := io.Copy(writer, io.LimitReader(stdout, maxDaemonLogSize))
	if err != nil {
		return err
	}
	if n >= maxDaemonLogSize {
		_, err = fmt.Fprintf(writer, "\n... (log truncated at 500 MB) ...\n")
		return err
	}
	return nil
}

// snapLogsMaxLines bounds how many journal entries snapd will return for the
// daemon's service. snapd's Logs API emits entries chronologically (oldest →
// newest within the kept window), so we can't byte-cap from the newest end the
// way `journalctl -r` allows; instead we cap on line count and let snapd drop
// the oldest entries beyond this window. At ~200 B/line this is ~20 MB —
// comfortably below the 500 MB diagnostics-archive cap.
const snapLogsMaxLines = 100000

// streamSnapLogs fetches the last snapLogsMaxLines entries of the named snap
// service's logs via snapd's privileged Logs API and writes them
// chronologically to writer. snapd is the only thing on the snap side with
// access to the host journal; sdjournal and journalctl run from inside snap
// confinement see only the snap's private mount-namespace view.
func streamSnapLogs(writer io.Writer, service string) error {
	c := client.New(nil)
	logs, err := c.Logs([]string{service}, client.LogOptions{
		N:      snapLogsMaxLines,
		Follow: false,
	})
	if err != nil {
		return err
	}

	var count int
	for entry := range logs {
		count++
		// nolint:errcheck
		fmt.Fprintf(writer, "%s %s[%s]: %s\n",
			entry.Timestamp.Format("Jan 02 15:04:05"),
			service, entry.PID, entry.Message)
	}
	if count == 0 {
		return fmt.Errorf("snapd returned no log entries for %s", service)
	}
	return nil
}

// streamFileToWriter streams the contents of filePath into writer, capped at
// maxDaemonLogSize bytes. For files within the cap the data is copied as-is
// (chronological). For oversized files the kept tail is emitted in reverse
// line order — newest first — so the file path matches the newest-first
// behaviour of the systemd `journalctl -r` path.
func streamFileToWriter(writer io.Writer, filePath string) error {
	f, err := os.Open(filePath) // #nosec G304 -- filePath comes from known system paths, not user input
	if err != nil {
		return err
	}
	defer f.Close() // nolint:errcheck

	info, err := f.Stat()
	if err != nil {
		return err
	}

	if info.Size() > maxDaemonLogSize {
		fmt.Fprintf(writer, "... (log truncated to last 500 MB, reversed) ...\n") // nolint:errcheck
		return writeFileTailReversed(writer, f, info.Size(), maxDaemonLogSize)
	}

	_, err = io.Copy(writer, f)
	return err
}

// writeFileTailReversed writes the last `max` bytes of f to w with line
// order reversed (last line first). Reads the file in 64 KiB chunks from the
// tail toward the truncation boundary so memory stays bounded by the longest
// line — never by the file or kept-tail size.
func writeFileTailReversed(w io.Writer, f *os.File, fileSize, max int64) error {
	const chunkSize int64 = 64 * 1024
	start := fileSize - max
	if start < 0 {
		start = 0
	}

	pos := fileSize
	// leftover holds bytes from earlier (toward BOF) than the chunks
	// processed so far but newer than the current read position — i.e. the
	// "head" of the previously-processed chunk that hadn't yet seen a
	// newline to its left.
	var leftover []byte

	emit := func(line []byte) error {
		if _, err := w.Write(line); err != nil {
			return err
		}
		_, err := w.Write([]byte{'\n'})
		return err
	}

	for pos > start {
		size := chunkSize
		if pos-size < start {
			size = pos - start
		}
		pos -= size
		chunk := make([]byte, size)
		if _, err := f.ReadAt(chunk, pos); err != nil {
			return err
		}

		lastNL := bytes.LastIndexByte(chunk, '\n')
		if lastNL == -1 {
			// No newline in this chunk: prepend whole chunk to leftover
			// (older bytes go in front of newer).
			leftover = append(append([]byte{}, chunk...), leftover...)
			continue
		}

		// Bytes after the last \n in chunk join with leftover to form one
		// complete line (the oldest unemitted line newer than this chunk).
		tail := chunk[lastNL+1:]
		if len(tail) > 0 || len(leftover) > 0 {
			line := append(append([]byte{}, tail...), leftover...)
			if err := emit(line); err != nil {
				return err
			}
		}

		// Walk remaining \n boundaries right-to-left, emitting each
		// complete line within this chunk.
		end := lastNL
		for {
			nl := bytes.LastIndexByte(chunk[:end], '\n')
			if nl == -1 {
				leftover = append([]byte{}, chunk[:end]...)
				break
			}
			if err := emit(chunk[nl+1 : end]); err != nil {
				return err
			}
			end = nl
		}
	}

	// Final leftover is the oldest line in the kept tail. May be partial if
	// the truncation boundary fell mid-line; that's acceptable since the
	// header already announced truncation.
	if len(leftover) > 0 {
		if err := emit(leftover); err != nil {
			return err
		}
	}
	return nil
}

func addDirectoryToZip(zipWriter *zip.Writer, dirPath, zipPrefix string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories themselves, only add files
		if info.IsDir() {
			return nil
		}

		// Skip symlinks to prevent the daemon from reading files
		// outside the intended directory via user-planted links.
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		// Get relative path from dirPath
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		zipPath := filepath.Join(zipPrefix, relPath)
		return addFileToZip(zipWriter, path, zipPath)
	})
}

func addFileToZip(zipWriter *zip.Writer, filePath, zipPath string) error {
	file, err := os.Open(filePath) // #nosec G304 -- symlinks are filtered by addDirectoryToZip before this is called
	if err != nil {
		return err
	}
	defer file.Close() // nolint:errcheck

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = zipPath
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

// writeBlock appends a titled section to w in the form:
//
//	=== title ===
//	<content>
//	=========
func writeBlock(w io.Writer, title, content string) {
	// nolint:errcheck
	{
		fmt.Fprintf(w, "=== %s ===\n", title)
		fmt.Fprint(w, content)
		fmt.Fprint(w, "=========\n\n")
	}
}

// runCommand executes name with args and returns the combined stdout/stderr.
// On failure the error is appended inline so the block still reports
// something meaningful instead of being silently empty.
func runCommand(name string, args ...string) string {
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return fmt.Sprintf("%serror: %v\n", out, err)
	}
	return string(out)
}

// readFile returns the contents of path as a string. On failure the error is
// rendered inline (mirroring runCommand) so blocks are never silently empty.
func readFile(path string) string {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return fmt.Sprintf("error reading %s: %v\n", path, err)
	}
	return string(data)
}

// dbusGetProperty fetches a property from the system bus and returns its
// formatted string representation. Errors are rendered inline (mirroring
// runCommand) so the block is never silently empty.
func dbusGetProperty(service, path, iface, property string) string {
	conn, err := dbus.SystemBus()
	if err != nil {
		return fmt.Sprintf("dbus connect error: %v\n", err)
	}
	v, err := conn.Object(service, dbus.ObjectPath(path)).GetProperty(iface + "." + property)
	if err != nil {
		return fmt.Sprintf("error: %v\n", err)
	}
	return v.String() + "\n"
}

func addSystemInfo(zipWriter *zip.Writer, state appState) error {
	w, err := zipWriter.Create("system-info.txt")
	if err != nil {
		return err
	}

	writeBlock(w, "OS Release", readFile("/etc/os-release"))

	if _, err := os.Stat("/etc/lsb-release"); err == nil {
		writeBlock(w, "Linux Distribution", readFile("/etc/lsb-release"))
	} else {
		writeBlock(w, "Linux Distribution", runCommand("lsb_release", "-a"))
	}

	writeBlock(w, "Kernel Version", runCommand("uname", "-a"))
	writeBlock(w, "Desktop Environment", collectDesktopEnvironment())

	// nordvpn version/status/settings come from in-process state pulled in
	// CollectDiagnostics — no shelling out to the CLI.
	writeBlock(w, "nordvpn version", state.version)
	writeBlock(w, "nordvpn status", state.status)
	writeBlock(w, "nordvpn settings", state.settings)

	return nil
}

// collectDesktopEnvironment returns per-session loginctl properties for each
// active session, formatted for inclusion in the system-info block.
func collectDesktopEnvironment() string {
	output, err := exec.Command("loginctl", "list-sessions", "--no-legend").Output()
	if err != nil {
		return fmt.Sprintf("loginctl error: %v\n", err)
	}
	var b strings.Builder
	for _, session := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		fields := strings.Fields(session)
		if len(fields) < 1 {
			continue
		}
		sessionID := fields[0]
		fmt.Fprintf(&b, "--- Session %s ---\n", sessionID)
		// #nosec G204 -- sessionID is safe to use
		if props, err := exec.Command("loginctl", "show-session", sessionID,
			"-p", "Type", "-p", "Desktop", "-p", "Remote", "-p", "User").Output(); err == nil {
			b.Write(props)
		}
	}
	return b.String()
}

func addNetworkInfo(zipWriter *zip.Writer) error {
	w, err := zipWriter.Create("network-info.txt")
	if err != nil {
		return err
	}

	writeBlock(w, "ip addr", runCommand("ip", "addr"))
	writeBlock(w, "ip rule show", runCommand("ip", "rule", "show"))
	writeBlock(w, "ip route show table all", runCommand("ip", "route", "show", "table", "all"))

	writeBlock(w, "net.ipv6.conf.*.disable_ipv6", readDisableIPv6Status())
	writeBlock(w, "net.ipv4.conf.all.rp_filter", readFile("/proc/sys/net/ipv4/conf/all/rp_filter"))

	return nil
}

// readDisableIPv6Status reads net.ipv6.conf.<iface>.disable_ipv6 for every
// interface (including "all" and "default") and returns a sysctl-style
// listing sorted by interface name.
func readDisableIPv6Status() string {
	matches, err := filepath.Glob("/proc/sys/net/ipv6/conf/*/disable_ipv6")
	if err != nil {
		return fmt.Sprintf("glob error: %v\n", err)
	}
	sort.Strings(matches)
	var b strings.Builder
	for _, m := range matches {
		iface := filepath.Base(filepath.Dir(m))
		data, err := os.ReadFile(m) // #nosec G304 -- m is from filepath.Glob with hardcoded /proc/sys pattern
		if err != nil {
			fmt.Fprintf(&b, "net.ipv6.conf.%s.disable_ipv6 = error: %v\n", iface, err)
			continue
		}
		fmt.Fprintf(&b, "net.ipv6.conf.%s.disable_ipv6 = %s", iface, data)
	}
	return b.String()
}

// addNFTablesInfo streams the full nftables ruleset into nftables-ruleset.txt.
// It lives in its own entry (rather than being a block inside network-info.txt)
// because the ruleset dump can be large enough to drown out the surrounding
// report.
func addNFTablesInfo(zipWriter *zip.Writer) error {
	w, err := zipWriter.Create("nftables-ruleset.txt")
	if err != nil {
		return err
	}
	return streamCommandToWriter(w, "nft", "list", "ruleset")
}

// addDNSInfo writes DNS-related diagnostics (resolv.conf, systemd-resolved,
// NetworkManager DNS state) to dns-info.txt inside the archive. DNS is a
// frequent support topic, so these blocks live in their own entry to keep
// them easy to find alongside the rest of the report.
func addDNSInfo(zipWriter *zip.Writer) error {
	w, err := zipWriter.Create("dns-info.txt")
	if err != nil {
		return err
	}

	writeBlock(w, "/etc/resolv.conf (ls -la)", runCommand("ls", "-la", "/etc/resolv.conf"))
	writeBlock(w, "/etc/resolv.conf", readFile("/etc/resolv.conf"))

	writeBlock(w, "systemd-resolve / resolvectl status", resolvectlStatus())

	writeBlock(w, "nmcli general", runCommand("nmcli", "general"))
	writeBlock(w, "nmcli device show", runCommand("nmcli", "device", "show"))

	writeBlock(w, "NetworkManager DNS Mode (dbus)", dbusGetProperty(
		"org.freedesktop.NetworkManager",
		"/org/freedesktop/NetworkManager/DnsManager",
		"org.freedesktop.NetworkManager.DnsManager", "Mode"))
	writeBlock(w, "NetworkManager DNS Configuration (dbus)", dbusGetProperty(
		"org.freedesktop.NetworkManager",
		"/org/freedesktop/NetworkManager/DnsManager",
		"org.freedesktop.NetworkManager.DnsManager", "Configuration"))

	writeBlock(w, "/etc/systemd/resolved.conf", readFile("/etc/systemd/resolved.conf"))

	// conf.d drop-ins land as real zip subdirectories so each file keeps its
	// name and can be inspected individually. A missing directory is logged
	// rather than skipped silently so support can tell the difference between
	// "no drop-ins configured" and "we forgot to collect them".
	if _, err := os.Stat("/etc/NetworkManager/conf.d"); err == nil {
		if err := addDirectoryToZip(zipWriter, "/etc/NetworkManager/conf.d", "etc/NetworkManager/conf.d"); err != nil {
			return err
		}
	} else {
		log.Info(logPrefix, "/etc/NetworkManager/conf.d:", err)
	}
	if _, err := os.Stat("/etc/systemd/resolved.conf.d"); err == nil {
		if err := addDirectoryToZip(zipWriter, "/etc/systemd/resolved.conf.d", "etc/systemd/resolved.conf.d"); err != nil {
			return err
		}
	} else {
		writeBlock(w, "/etc/systemd/resolved.conf.d", fmt.Sprintf("error: %v\n", err))
	}

	return nil
}

// resolvectlStatus runs the resolver status command, preferring the legacy
// systemd-resolve binary when present (old systems) and falling back to
// resolvectl (new systems). Returns the combined output.
func resolvectlStatus() string {
	if _, err := exec.LookPath("systemd-resolve"); err == nil {
		return runCommand("systemd-resolve", "--status")
	}
	return runCommand("resolvectl", "status")
}
