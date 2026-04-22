package daemon

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/snapconf"
)

const (
	// maxDaemonLogSize is the maximum size of daemon logs to collect (1GB)
	maxDaemonLogSize = 1 * 1024 * 1024 * 1024

	// maxZipFileSize caps the resulting diagnostics archive (200 MB).
	maxZipFileSize = 200 * 1024 * 1024

	zipTooLargeMsg = "Diagnostics file exceeds 200 MB limit. Please contact support for assistance."
)

// errZipSizeLimitExceeded is returned when writing more data would push the
// diagnostics zip past maxZipFileSize.
var errZipSizeLimitExceeded = errors.New("diagnostics zip exceeds size limit")

// sizeLimitedWriter wraps an io.Writer and rejects writes that would exceed
// the configured limit, returning errZipSizeLimitExceeded.
type sizeLimitedWriter struct {
	w       io.Writer
	limit   int64
	written int64
}

func (lw *sizeLimitedWriter) Write(p []byte) (int, error) {
	// Subtraction avoids int64 overflow that a naive `written+len(p) > limit`
	// could hit when written is close to math.MaxInt64.
	if int64(len(p)) > lw.limit-lw.written {
		return 0, errZipSizeLimitExceeded
	}
	n, err := lw.w.Write(p)
	lw.written += int64(n)
	return n, err
}

func (r *RPC) CollectDiagnostics(in *pb.Empty, srv pb.Daemon_CollectDiagnosticsServer) (retErr error) {
	caller, err := resolveDiagnosticsCaller(srv.Context())
	if err != nil {
		log.Println(internal.ErrorPrefix, err)
		return srv.Send(&pb.DiagnosticsProgress{Error: err.Error()})
	}

	srv.Send(&pb.DiagnosticsProgress{Percentage: 0, Step: "Initializing..."})

	zipFile, err := os.Create(caller.zipPath)
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to create diagnostics zip:", err)
		return srv.Send(&pb.DiagnosticsProgress{Error: fmt.Sprintf("Failed to create zip file: %v", err)})
	}
	defer func() {
		zipFile.Close()
		if retErr != nil {
			os.Remove(caller.zipPath)
		}
	}()

	// Change ownership immediately so user can access partial file
	if err := os.Chown(caller.zipPath, int(caller.uid), int(caller.gid)); err != nil {
		log.Println(internal.WarningPrefix, "failed to change file ownership:", err)
	}

	if err := collectDiagnosticsData(srv, zipFile, caller.user.HomeDir, r.version); err != nil {
		if errors.Is(err, errZipSizeLimitExceeded) {
			log.Println(internal.ErrorPrefix, "diagnostics zip exceeded 200 MB limit")
			srv.Send(&pb.DiagnosticsProgress{Error: zipTooLargeMsg})
			return err
		}
		log.Println(internal.ErrorPrefix, "failed to collect diagnostics:", err)
		srv.Send(&pb.DiagnosticsProgress{Error: fmt.Sprintf("Failed to collect diagnostics: %v", err)})
		return err
	}

	if err := zipFile.Close(); err != nil {
		log.Println(internal.ErrorPrefix, "failed to close zip file:", err)
		srv.Send(&pb.DiagnosticsProgress{Error: fmt.Sprintf("Failed to close zip file: %v", err)})
		return err
	}

	return srv.Send(&pb.DiagnosticsProgress{
		Percentage: 100,
		Step:       "Done",
		Done:       true,
		FilePath:   caller.zipPath,
	})
}

// diagnosticsStep represents one unit of diagnostics collection: the message
// shown to the user and the function that performs the work. Progress
// percentages are computed dynamically from the step's position in the list.
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
// step, reporting progress to the client. Individual step failures are logged
// as warnings and do not abort collection — partial diagnostics are still
// useful. Size-limit overflow is fatal and is returned to the caller so the
// partial zip can be discarded.
//
// Progress percentages are distributed evenly across steps: step i (0-indexed)
// is reported at (i+1) * 100 / (N+1), leaving the final 100% slot for "Done".
func collectDiagnosticsData(
	srv pb.Daemon_CollectDiagnosticsServer,
	output io.Writer,
	homeDir, version string,
) error {
	limited := &sizeLimitedWriter{w: output, limit: maxZipFileSize}
	zipWriter := zip.NewWriter(limited)
	defer zipWriter.Close()

	steps := []diagnosticsStep{
		{"Collecting daemon logs...", func() error {
			return addDaemonLogs(zipWriter)
		}, true},
		{"Collecting CLI logs...", func() error {
			cliLog := filepath.Join(homeDir, ".config", "nordvpn", "cli.log")
			return addFileToZip(zipWriter, cliLog, "cli.log")
		}, false},
		{"Collecting user logs...", func() error {
			cacheDir := filepath.Join(homeDir, ".cache", "nordvpn")
			return addDirectoryToZip(zipWriter, cacheDir, "cache")
		}, false},
		{"Collecting system info...", func() error {
			return addSystemInfo(zipWriter)
		}, false},
		{"Collecting version info...", func() error {
			return addVersionInfo(zipWriter, version)
		}, false},
		{"Collecting network info...", func() error {
			return addNetworkInfo(zipWriter)
		}, false},
	}

	slots := int32(len(steps) + 1) // +1 reserves the final slot for "Done" at 100%.
	for i, step := range steps {
		pct := int32(i+1) * 100 / slots
		srv.Send(&pb.DiagnosticsProgress{Percentage: pct, Step: step.description})
		if err := step.collect(); err != nil {
			if step.fatal || errors.Is(err, errZipSizeLimitExceeded) {
				return err
			}
			log.Println(internal.WarningPrefix, "diagnostics step failed:", step.description, err)
		}
	}

	// Explicit finalize so caller sees any central-directory write error
	// (e.g. hitting the size cap on the last flush).
	return zipWriter.Close()
}

// addDaemonLogs writes the daemon's logs to daemon.log inside the archive.
// It probes how the daemon is being supervised (snap > systemd > initd) and
// pulls logs from the matching source. If none of these apply, an explanatory
// note is written into the archive entry instead.
func addDaemonLogs(zipWriter *zip.Writer) error {
	writer, err := zipWriter.Create("daemon.log")
	if err != nil {
		return err
	}

	switch {
	case snapconf.IsUnderSnap():
		if err := streamCommandToWriter(writer, "snap", "logs", "nordvpn", "-n", "all"); err != nil {
			fmt.Fprintf(writer, "Error getting snap logs: %v\n", err)
		}
	case internal.IsSystemd():
		if err := streamCommandToWriter(writer, "journalctl", "-u", "nordvpnd", "--no-pager"); err != nil {
			fmt.Fprintf(writer, "Error getting journalctl output: %v\n", err)
		}
	case internal.FileExists("/etc/init.d/nordvpn"):
		logFile := filepath.Join(internal.LogPath, "daemon.log")
		if err := streamFileToWriter(writer, logFile); err != nil {
			fmt.Fprintf(writer, "Error reading %s: %v\n", logFile, err)
		}
	default:
		return fmt.Errorf("unable to determine how the daemon is running (systemd/snap/initd)")
	}

	return nil
}

// streamCommandToWriter runs a command and streams its output directly to the writer
// with a maximum of maxDaemonLogSize bytes
func streamCommandToWriter(writer io.Writer, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// Copy in chunks, tracking total written
	buf := make([]byte, 32*1024) // 32KB buffer
	var totalWritten int64

	for totalWritten < maxDaemonLogSize {
		n, readErr := stdout.Read(buf)
		if n > 0 {
			// Check if this chunk would exceed the limit
			remaining := maxDaemonLogSize - totalWritten
			toWrite := int64(n)
			if toWrite > remaining {
				toWrite = remaining
			}

			written, writeErr := writer.Write(buf[:toWrite])
			totalWritten += int64(written)
			if writeErr != nil {
				cmd.Process.Kill()
				cmd.Wait()
				return writeErr
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			cmd.Process.Kill()
			cmd.Wait()
			return readErr
		}
	}

	// If we hit the limit, kill the command and note truncation
	if totalWritten >= maxDaemonLogSize {
		cmd.Process.Kill()
		fmt.Fprintf(writer, "\n... (log truncated at 1GB) ...\n")
	}

	cmd.Wait()
	return nil
}

// streamFileToWriter streams a file directly to the writer with a maximum of maxDaemonLogSize bytes
func streamFileToWriter(writer io.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Check file size to see if we need to skip to the end
	info, err := file.Stat()
	if err != nil {
		return err
	}

	// If file is larger than max, seek to get the last maxDaemonLogSize bytes
	if info.Size() > maxDaemonLogSize {
		fmt.Fprintf(writer, "... (log truncated to last 1GB) ...\n")
		if _, err := file.Seek(-maxDaemonLogSize, io.SeekEnd); err != nil {
			return err
		}
	}

	_, err = io.Copy(writer, file)
	return err
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
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

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
//	<content produced by fn>
//	=========
//
// The content is produced by fn, which writes directly to the block body so
// large outputs can stream without being buffered in memory first. If fn
// returns an error it is rendered inline; the header and footer are still
// emitted so the surrounding report stays well-formed.
func writeBlock(w io.Writer, title string, fn func(io.Writer) error) {
	fmt.Fprintf(w, "=== %s ===\n", title)
	if err := fn(w); err != nil {
		fmt.Fprintf(w, "(error: %v)\n", err)
	}
	fmt.Fprint(w, "=========\n")
}

// blockString adapts a ready-made string into a writeBlock content producer.
func blockString(s string) func(io.Writer) error {
	return func(w io.Writer) error {
		_, err := io.WriteString(w, s)
		return err
	}
}

// blockCommand adapts `name args...` into a writeBlock content producer that
// streams the command's stdout directly into the block.
func blockCommand(name string, args ...string) func(io.Writer) error {
	return func(w io.Writer) error {
		return streamCommandToWriter(w, name, args...)
	}
}

func addSystemInfo(zipWriter *zip.Writer) error {
	w, err := zipWriter.Create("system-info.txt")
	if err != nil {
		return err
	}

	osRelease, _ := os.ReadFile("/etc/os-release")
	writeBlock(w, "OS Release", blockString(string(osRelease)))

	if distro, err := os.ReadFile("/etc/lsb-release"); err == nil {
		writeBlock(w, "Linux Distribution", blockString(string(distro)))
	} else {
		writeBlock(w, "Linux Distribution", blockCommand("lsb_release", "-a"))
	}

	writeBlock(w, "Kernel Version", blockCommand("uname", "-a"))
	writeBlock(w, "Desktop Environment", blockString(collectDesktopEnvironment()))
	writeBlock(w, "Systemd Status", blockCommand("systemctl", "status", "nordvpnd", "--no-pager"))

	return nil
}

// collectDesktopEnvironment returns per-session loginctl properties for each
// active session, formatted for inclusion in the system-info block.
func collectDesktopEnvironment() string {
	output, err := exec.Command("loginctl", "list-sessions", "--no-legend").Output()
	if err != nil {
		return ""
	}
	var b strings.Builder
	for _, session := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		fields := strings.Fields(session)
		if len(fields) < 1 {
			continue
		}
		sessionID := fields[0]
		fmt.Fprintf(&b, "--- Session %s ---\n", sessionID)
		if props, err := exec.Command("loginctl", "show-session", sessionID,
			"-p", "Type", "-p", "Desktop", "-p", "Remote", "-p", "User").Output(); err == nil {
			b.Write(props)
		}
	}
	return b.String()
}

func addVersionInfo(zipWriter *zip.Writer, version string) error {
	w, err := zipWriter.Create("version-info.txt")
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "NordVPN Version: %s\n", version)
	fmt.Fprintf(w, "Collection Time: %s\n", time.Now().Format(time.RFC3339))
	return nil
}

func addNetworkInfo(zipWriter *zip.Writer) error {
	w, err := zipWriter.Create("network-info.txt")
	if err != nil {
		return err
	}

	writeBlock(w, "Network Interfaces", blockCommand("ip", "addr"))
	writeBlock(w, "IP Rules", blockCommand("ip", "rule", "show"))
	writeBlock(w, "Routing Tables", blockCommand("ip", "route", "show", "table", "all"))

	writeBlock(w, "/etc/resolv.conf", streamFileWithMetadata("/etc/resolv.conf"))

	// systemd-resolve was renamed to resolvectl; try the new name first and
	// fall back so both old and new systems are covered in a single block.
	writeBlock(w, "systemd-resolve / resolvectl status",
		blockCommand(resolvectlCmd(), resolvectlArgs()...))

	writeBlock(w, "nmcli general", blockCommand("nmcli", "general"))
	writeBlock(w, "nmcli device show (DNS)", blockString(dnsLinesFromNmcli()))

	writeBlock(w, "NetworkManager DNS Mode (busctl)", blockCommand(
		"busctl", "get-property", "org.freedesktop.NetworkManager",
		"/org/freedesktop/NetworkManager/DnsManager",
		"org.freedesktop.NetworkManager.DnsManager", "Mode"))
	writeBlock(w, "NetworkManager DNS Configuration (busctl)", blockCommand(
		"busctl", "get-property", "org.freedesktop.NetworkManager",
		"/org/freedesktop/NetworkManager/DnsManager",
		"org.freedesktop.NetworkManager.DnsManager", "Configuration"))

	writeBlock(w, "/etc/NetworkManager/conf.d/", blockString(dumpConfDir("/etc/NetworkManager/conf.d")))

	resolvedConf, _ := os.ReadFile("/etc/systemd/resolved.conf")
	writeBlock(w, "/etc/systemd/resolved.conf", blockString(string(resolvedConf)))

	writeBlock(w, "/etc/systemd/resolved.conf.d/", blockString(dumpConfDir("/etc/systemd/resolved.conf.d")))

	writeBlock(w, "NFTables Ruleset", blockCommand("nft", "list", "ruleset"))

	return nil
}

// resolvectlCmd picks systemd-resolve (old) if present, otherwise resolvectl.
func resolvectlCmd() string {
	if _, err := exec.LookPath("systemd-resolve"); err == nil {
		return "systemd-resolve"
	}
	return "resolvectl"
}

func resolvectlArgs() []string {
	if _, err := exec.LookPath("systemd-resolve"); err == nil {
		return []string{"--status"}
	}
	return []string{"status"}
}

// dnsLinesFromNmcli runs `nmcli device show` and returns only the lines
// relevant to DNS debugging (device identity + DNS entries).
func dnsLinesFromNmcli() string {
	output, err := exec.Command("nmcli", "device", "show").Output()
	if err != nil {
		return fmt.Sprintf("(nmcli error: %v)\n", err)
	}
	var b strings.Builder
	for _, line := range strings.Split(string(output), "\n") {
		if strings.Contains(line, "DNS") ||
			strings.Contains(line, "DEVICE") ||
			strings.Contains(line, "TYPE") ||
			strings.Contains(line, "CONNECTION") {
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// streamFileWithMetadata returns a writeBlock content producer that streams
// `ls -la` output for path (which exposes size, permissions, and symlink
// targets) followed by the file's contents straight into the block.
func streamFileWithMetadata(path string) func(io.Writer) error {
	return func(w io.Writer) error {
		if err := streamCommandToWriter(w, "ls", "-la", path); err != nil {
			fmt.Fprintf(w, "(error running ls: %v)\n", err)
		}
		return streamFileToWriter(w, path)
	}
}

// dumpConfDir concatenates every regular file in dir as `-- name --` sections.
func dumpConfDir(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "(directory not found)\n"
	}
	var b strings.Builder
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fmt.Fprintf(&b, "-- %s --\n", entry.Name())
		if data, err := os.ReadFile(filepath.Join(dir, entry.Name())); err == nil {
			b.Write(data)
			b.WriteByte('\n')
		}
	}
	return b.String()
}
