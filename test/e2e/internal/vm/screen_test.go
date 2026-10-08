package vm

import (
	"image"
	"os/exec"
	"testing"
)

const tesseractTSV = "level\tpage_num\tblock_num\tpar_num\tline_num\tword_num\tleft\ttop\twidth\theight\tconf\ttext\n" +
	"1\t1\t0\t0\t0\t0\t0\t0\t800\t600\t-1\t\n" +
	"5\t1\t1\t1\t1\t1\t100\t40\t60\t20\t96.1\tNot\n" +
	"5\t1\t1\t1\t1\t2\t170\t40\t90\t20\t95.0\tsecured\n" +
	"5\t1\t2\t1\t1\t1\t200\t240\t80\t24\t91.2\tSecure\n" +
	"5\t1\t2\t1\t1\t2\t290\t240\t30\t24\t90.0\tmy\n" +
	"5\t1\t2\t1\t1\t3\t330\t240\t140\t24\t92.5\tconnection\n" +
	"5\t1\t3\t1\t1\t1\t200\t300\t80\t24\t12.0\tgarbage\n" +
	"5\t1\t4\t1\t1\t1\t200\t400\t80\t24\t90.0\tSecure\n" +
	"5\t1\t4\t1\t2\t1\t200\t430\t30\t24\t90.0\tmy\n" +
	"5\t1\t4\t1\t2\t2\t240\t430\t140\t24\t90.0\tconnection\n"

func TestParseTSV(t *testing.T) {
	words := parseTSV([]byte(tesseractTSV))
	// All words but the low confidence one.
	if len(words) != 8 {
		t.Fatalf("got %d words, want 8: %v", len(words), words)
	}
	if got, want := words[2].Box, image.Rect(200, 240, 280, 264); got != want {
		t.Errorf("box of %q = %v, want %v", words[2].Text, got, want)
	}
}

func TestFindText(t *testing.T) {
	words := parseTSV([]byte(tesseractTSV))
	tests := []struct {
		phrase string
		want   image.Rectangle
		found  bool
	}{
		{phrase: "Secure my connection", want: image.Rect(200, 240, 470, 264), found: true},
		{phrase: "secure MY connection.", want: image.Rect(200, 240, 470, 264), found: true},
		{phrase: "Not secured", want: image.Rect(100, 40, 260, 60), found: true},
		{phrase: "Notsecured", want: image.Rect(100, 40, 260, 60), found: true},
		{phrase: "connection", want: image.Rect(330, 240, 470, 264), found: true},
		{phrase: "garbage"},
		{phrase: "Connected"},
	}
	for _, tt := range tests {
		got, found := FindText([][]Word{words}, tt.phrase)
		if found != tt.found || got != tt.want {
			t.Errorf("FindText(%q) = %v, %v, want %v, %v", tt.phrase, got, found, tt.want, tt.found)
		}
	}
}

func TestFindTextPrefersWholeLine(t *testing.T) {
	words := []Word{
		{Text: "By", Box: image.Rect(0, 0, 10, 10), line: "1"},
		{Text: "selecting", Box: image.Rect(10, 0, 50, 10), line: "1"},
		{Text: "Accept,", Box: image.Rect(50, 0, 80, 10), line: "1"},
		{Text: "you", Box: image.Rect(80, 0, 100, 10), line: "1"},
		{Text: "Accept", Box: image.Rect(200, 100, 240, 110), line: "2"},
	}
	box, found := FindText([][]Word{words}, "Accept")
	if want := image.Rect(200, 100, 240, 110); !found || box != want {
		t.Errorf("FindText = %v, %v, want the button at %v", box, found, want)
	}
	// Without a whole line match, the phrase inside a text is good too.
	box, found = FindText([][]Word{words[:4]}, "Accept")
	if want := image.Rect(50, 0, 80, 10); !found || box != want {
		t.Errorf("FindText = %v, %v, want %v", box, found, want)
	}
}

func TestFindTextRequiresOneLine(t *testing.T) {
	// "Secure" and "my connection" on different lines of block 4 only.
	words := parseTSV([]byte(tesseractTSV))[5:]
	if box, found := FindText([][]Word{words}, "Secure my connection"); found {
		t.Errorf("found phrase split over lines at %v", box)
	}
}

func TestReadTextFedoraScreenshot(t *testing.T) {
	if _, err := exec.LookPath("tesseract"); err != nil {
		t.Skip("tesseract is not installed")
	}
	img := loadPNG(t, "testdata/fedora-privacy-dialog-and-tray-menu.png")
	passes, dump, err := readText(img, img.Bounds(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		phrase string
		center image.Point // where the text is on the screenshot
	}{
		{phrase: "We value your privacy", center: image.Pt(530, 292)},
		{phrase: "Reject non-essential", center: image.Pt(640, 596)},
		{phrase: "Accept", center: image.Pt(830, 596)},            // white on blue
		{phrase: "Log in", center: image.Pt(1076, 212)},           // light on dark
		{phrase: "Open NordVPN app", center: image.Pt(1123, 119)}, // light on dark
	}
	for _, tt := range tests {
		box, found := FindText(passes, tt.phrase)
		if !found {
			t.Errorf("%q not found, OCR:\n%s", tt.phrase, dump)
			continue
		}
		if c := Center(box); abs(c.X-tt.center.X) > 15 || abs(c.Y-tt.center.Y) > 10 {
			t.Errorf("%q found at %v, want around %v", tt.phrase, c, tt.center)
		}
	}
}

func abs(x int) int { return max(x, -x) }
