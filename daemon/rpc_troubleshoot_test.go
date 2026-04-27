package daemon

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// mockDiagnosticsServer captures DiagnosticsProgress messages sent by the RPC
// so tests can assert on the stream contents.
type mockDiagnosticsServer struct {
	grpc.ServerStream
	ctx  context.Context
	msgs []*pb.DiagnosticsProgress
}

func (m *mockDiagnosticsServer) Send(p *pb.DiagnosticsProgress) error {
	m.msgs = append(m.msgs, p)
	return nil
}

func (m *mockDiagnosticsServer) Context() context.Context {
	if m.ctx == nil {
		return context.Background()
	}
	return m.ctx
}

func TestSizeLimitedWriter(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name        string
		limit       int64
		writes      [][]byte
		accepted    []int
		writeErrors map[int]bool
		expectBuf   string // cumulative bytes that should reach the underlying writer
	}{
		{
			name:      "under limit",
			limit:     10,
			writes:    [][]byte{[]byte("hello")},
			accepted:  []int{5},
			expectBuf: "hello",
		},
		{
			name:      "exact limit",
			limit:     5,
			writes:    [][]byte{[]byte("hello")},
			accepted:  []int{5},
			expectBuf: "hello",
		},
		{
			name:        "single write over limit truncates to remaining",
			limit:       4,
			writes:      [][]byte{[]byte("hello")},
			accepted:    []int{4},
			writeErrors: map[int]bool{0: true},
			expectBuf:   "hell",
		},
		{
			name:        "cumulative overflow truncates the overflowing write",
			limit:       6,
			writes:      [][]byte{[]byte("abc"), []byte("defg")},
			accepted:    []int{3, 3},
			writeErrors: map[int]bool{1: true},
			expectBuf:   "abcdef",
		},
		{
			name:      "cumulative under limit",
			limit:     10,
			writes:    [][]byte{[]byte("abc"), []byte("defg")},
			accepted:  []int{3, 4},
			expectBuf: "abcdefg",
		},
		{
			name:        "subsequent write after limit hit returns no bytes",
			limit:       3,
			writes:      [][]byte{[]byte("abc"), []byte("d")},
			accepted:    []int{3, 0},
			writeErrors: map[int]bool{1: true},
			expectBuf:   "abc",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			lw := &sizeLimitedWriter{w: &buf, limit: tc.limit}
			for i, data := range tc.writes {
				n, err := lw.Write(data)
				assert.Equal(t, tc.accepted[i], n, "call %d accepted bytes", i)
				if tc.writeErrors[i] {
					assert.ErrorIs(t, err, errZipSizeLimitExceeded, "call %d expected size-limit error", i)
				} else {
					assert.NoError(t, err, "call %d unexpected error", i)
				}
			}
			assert.Equal(t, tc.expectBuf, buf.String(), "underlying buffer contents")
		})
	}
}

// TestSizeLimitedWriter_OverflowGuard verifies the subtraction-based bounds
// check in Write — a naive `written + len(p) > limit` would falsely accept
// writes when `written` is near math.MaxInt64.
func TestSizeLimitedWriter_OverflowGuard(t *testing.T) {
	category.Set(t, category.Unit)

	var buf bytes.Buffer
	lw := &sizeLimitedWriter{
		w:       &buf,
		limit:   1 << 30,
		written: 1 << 30, // already at limit
	}
	n, err := lw.Write([]byte("x"))
	assert.Equal(t, 0, n)
	assert.ErrorIs(t, err, errZipSizeLimitExceeded)
}

func TestWriteBlock(t *testing.T) {
	category.Set(t, category.Unit)

	var buf bytes.Buffer
	writeBlock(&buf, "My Title", "line1\nline2\n")
	assert.Equal(t, "=== My Title ===\nline1\nline2\n=========\n\n", buf.String())
}

func TestReadFile(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name           string
		setup          func(t *testing.T) string // returns path to read
		expectedOutput string
	}{
		{
			name: "present file returns its contents",
			setup: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), uuid.NewString()+".txt")
				require.NoError(t, os.WriteFile(path, []byte("hello"), 0600))
				return path
			},
			expectedOutput: "hello",
		},
		{
			name: "missing file returns inline error",
			setup: func(t *testing.T) string {
				return "/nonexistent/path/xyz.txt"
			},
			expectedOutput: "error reading /nonexistent/path/xyz.txt:",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := tc.setup(t)
			assert.Contains(t, readFile(path), tc.expectedOutput)
		})
	}
}

func TestRunCommand(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name           string
		cmd            string
		args           []string
		expectedOutput string
	}{
		{
			name:           "successful command returns its output",
			cmd:            "echo",
			args:           []string{"hi"},
			expectedOutput: "hi\n",
		},
		{
			name:           "failing command appends inline error",
			cmd:            "false",
			expectedOutput: "error:",
		},
		{
			name:           "missing binary appends inline error",
			cmd:            "/nonexistent/command-xyz",
			expectedOutput: "error:",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Contains(t, runCommand(tc.cmd, tc.args...), tc.expectedOutput)
		})
	}
}

func TestAddFileToZip(t *testing.T) {
	category.Set(t, category.Unit)

	srcPath := filepath.Join(t.TempDir(), uuid.NewString()+".txt")
	require.NoError(t, os.WriteFile(srcPath, []byte("file-content"), 0600))

	zipEntry := "dest/" + uuid.NewString() + ".txt"
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	require.NoError(t, addFileToZip(zw, srcPath, zipEntry))
	require.NoError(t, zw.Close())

	entries := readZipEntries(t, buf.Bytes())
	require.Contains(t, entries, zipEntry)
	assert.Equal(t, "file-content", entries[zipEntry])
}

func TestAddFileToZip_Missing(t *testing.T) {
	category.Set(t, category.Unit)

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	assert.Error(t, addFileToZip(zw, "/nonexistent/abc", "e.txt"))
}

func TestAddDirectoryToZip(t *testing.T) {
	category.Set(t, category.Unit)

	dir := t.TempDir()
	subName := uuid.NewString()
	sub := filepath.Join(dir, subName)
	require.NoError(t, os.MkdirAll(sub, 0700))

	fileA := uuid.NewString() + ".txt"
	fileB := uuid.NewString() + ".txt"
	require.NoError(t, os.WriteFile(filepath.Join(dir, fileA), []byte("aa"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(sub, fileB), []byte("bb"), 0600))

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	require.NoError(t, addDirectoryToZip(zw, dir, "prefix"))
	require.NoError(t, zw.Close())

	entries := readZipEntries(t, buf.Bytes())
	assert.Equal(t, "aa", entries["prefix/"+fileA])
	assert.Equal(t, "bb", entries["prefix/"+subName+"/"+fileB])
	assert.Len(t, entries, 2)
}

func TestAddDirectoryToZip_Missing(t *testing.T) {
	category.Set(t, category.Unit)

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	assert.Error(t, addDirectoryToZip(zw, "/nonexistent/xyz", "p"))
}

func TestWriteLogExtractionReport(t *testing.T) {
	category.Set(t, category.Unit)

	var zbuf bytes.Buffer
	zw := zip.NewWriter(&zbuf)

	logBuf := bytes.NewBufferString("line1\nline2\n")
	writeLogExtractionReport(zw, logBuf)
	require.NoError(t, zw.Close())

	entries := readZipEntries(t, zbuf.Bytes())
	assert.Equal(t, "line1\nline2\n", entries["log_extraction_report.log"])
}

func TestStreamFileToWriter_Small(t *testing.T) {
	category.Set(t, category.Unit)

	path := filepath.Join(t.TempDir(), uuid.NewString()+".log")
	require.NoError(t, os.WriteFile(path, []byte("small log"), 0600))

	var buf bytes.Buffer
	require.NoError(t, streamFileToWriter(&buf, path))
	assert.Equal(t, "small log", buf.String())
}

func TestStreamFileToWriter_Missing(t *testing.T) {
	category.Set(t, category.Unit)

	var buf bytes.Buffer
	assert.Error(t, streamFileToWriter(&buf, "/nonexistent/x.log"))
}

func TestStreamCommandToWriter(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name           string
		cmd            string
		args           []string
		expectErr      bool
		expectedOutput string
	}{
		{
			name:           "successful command streams output",
			cmd:            "echo",
			args:           []string{"hi"},
			expectedOutput: "hi\n",
		},
		{
			name:      "missing binary returns error",
			cmd:       "/nonexistent/command-xyz",
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := streamCommandToWriter(&buf, tc.cmd, tc.args...)
			if tc.expectErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedOutput, buf.String())
		})
	}
}

func TestFailToExtractDaemonLogs(t *testing.T) {
	category.Set(t, category.Unit)

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	err := addDaemonLogs(zw, daemonSupervisorUnknown)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to extract daemon logs automatically")
	assert.Contains(t, err.Error(), "systemd or snap")
	assert.Contains(t, err.Error(), "contact customer support")
}

// TestCreateDiagnosticsZip_UniqueFilename verifies that back-to-back calls
// produce distinct paths even when they share the same second-precision
// timestamp — the random suffix injected by os.CreateTemp's `*` is what
// guarantees uniqueness, so two collections in the same second cannot
// clobber each other's output.
func TestCreateDiagnosticsZip_UniqueFilename(t *testing.T) {
	category.Set(t, category.Unit)

	dir := t.TempDir()

	const N = 5
	seen := make(map[string]bool, N)
	for i := 0; i < N; i++ {
		f, err := createDiagnosticsZip(dir)
		require.NoError(t, err)
		require.NoError(t, f.Close())

		name := filepath.Base(f.Name())
		assert.False(t, seen[name], "duplicate filename %q (iteration %d)", name, i)
		seen[name] = true

		assert.True(t, strings.HasPrefix(name, "nordvpn-diagnostics-"), "unexpected prefix: %q", name)
		assert.True(t, strings.HasSuffix(name, ".zip"), "unexpected suffix: %q", name)
	}
	assert.Len(t, seen, N, "expected %d unique filenames", N)
}

func TestTroubleshootFailsWhengRPCPeerIsInvalid(t *testing.T) {
	category.Set(t, category.Unit)

	srv := &mockDiagnosticsServer{ctx: context.Background()}
	rpc := &RPC{version: "test"}

	err := rpc.CollectDiagnostics(&pb.Empty{}, srv)
	// Send itself succeeds, so the return value is nil; the error surfaces
	// as a populated Error field on the sent message.
	assert.NoError(t, err)
	require.Len(t, srv.msgs, 1)
	assert.NotEmpty(t, srv.msgs[0].Error)
	assert.Empty(t, srv.msgs[0].FilePath)
}

// readZipEntries opens the zip bytes and returns a map of name → contents for
// all file entries. Fails the test on any read error.
func readZipEntries(t *testing.T, data []byte) map[string]string {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)

	out := map[string]string{}
	for _, f := range zr.File {
		rc, err := f.Open()
		require.NoError(t, err)
		content, err := io.ReadAll(rc)
		rc.Close()
		require.NoError(t, err)
		out[f.Name] = string(content)
	}
	return out
}
