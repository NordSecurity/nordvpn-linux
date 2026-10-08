package vm

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode"

	"golang.org/x/image/draw"
)

const (
	pollInterval      = 500 * time.Millisecond
	stableDiff        = 0.001
	ocrScale          = 2
	minWordConfidence = 40
	// lightTextMaxBackground is 175 so that text near a button's rounded corner
	// still qualifies: the adaptive window partially overlaps the white card
	// behind the button, pushing the local mean to ~160–170, well below a plain
	// white background (~230+).
	lightTextWindow        = 41
	lightTextContrast      = 25
	lightTextMaxBackground = 175
)

func (m *Machine) Screenshot(t *testing.T) image.Image {
	t.Helper()
	path := filepath.Join(m.dir, "screen.png")
	if err := m.qmp.screendump(path); err != nil {
		t.Fatalf("taking screenshot: %v", err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("opening screenshot: %v", err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("decoding screenshot: %v", err)
	}
	m.lastScreenshot = path
	m.screen = img.Bounds()
	return img
}

func (m *Machine) WaitForStableScreen(t *testing.T, timeout time.Duration) image.Image {
	t.Helper()
	deadline := time.Now().Add(timeout)
	prev := m.Screenshot(t)
	for {
		time.Sleep(pollInterval)
		cur := m.Screenshot(t)
		if diffRatio(prev, cur) <= stableDiff {
			return cur
		}
		if time.Now().After(deadline) {
			t.Fatalf("screen didn't stop changing within %s", timeout)
		}
		prev = cur
	}
}

func diffRatio(a, b image.Image) float64 {
	if a.Bounds() != b.Bounds() {
		return 1
	}
	r := a.Bounds()
	diff := 0
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			if a.At(x, y) != b.At(x, y) {
				diff++
			}
		}
	}
	return float64(diff) / float64(r.Dx()*r.Dy())
}

type Word struct {
	Text string
	Box  image.Rectangle // in screen coordinates
	line string          // words with the same line are on the same text line
}

func (m *Machine) ReadText(t *testing.T, img image.Image, region image.Rectangle) [][]Word {
	t.Helper()
	passes, dump, err := readText(img, region, m.dir)
	if err != nil {
		t.Fatalf("reading text: %v", err)
	}
	m.lastOCR = dump
	return passes
}

// readText recognizes the words in region of img and returns them per OCR
// pass, plus a dump of the recognized text for debugging. workDir is where
// temporary images go.
//
// Tesseract reads dark text on light background best, but UIs have all
// kinds, so the region is read in several passes:
//   - as is, for dark text on light background;
//   - inverted, for light text on dark background (menus, dark themes);
//   - light text only: pixels lighter than their surroundings become black
//     text on white. For white text on colored buttons, which tesseract
//     otherwise skips as pictures.
func readText(img image.Image, region image.Rectangle, workDir string) ([][]Word, string, error) {
	region = region.Intersect(img.Bounds())
	gray := image.NewGray(image.Rect(0, 0, region.Dx()*ocrScale, region.Dy()*ocrScale))
	draw.CatmullRom.Scale(gray, gray.Bounds(), img, region, draw.Src, nil)

	inverted := image.NewGray(gray.Bounds())
	for i, v := range gray.Pix {
		inverted.Pix[i] = 255 - v
	}
	lightText := lightText(gray)

	var passes [][]Word
	var dump strings.Builder
	for i, pass := range []*image.Gray{gray, inverted, lightText} {
		words, err := ocr(pass, workDir)
		if err != nil {
			return nil, "", err
		}
		for j := range words {
			b := words[j].Box
			words[j].Box = image.Rect(b.Min.X/ocrScale, b.Min.Y/ocrScale, b.Max.X/ocrScale, b.Max.Y/ocrScale).
				Add(region.Min)
		}
		passes = append(passes, words)
		fmt.Fprintf(&dump, "pass %d (region %v): %s\n", i, region, joinWords(words))
	}
	return passes, dump.String(), nil
}

func lightText(g *image.Gray) *image.Gray {
	w, h := g.Bounds().Dx(), g.Bounds().Dy()
	// sum[y][x] is the sum of the pixels above and left of (x, y).
	sum := make([]int, (w+1)*(h+1))
	for y := 0; y < h; y++ {
		row := 0
		for x := 0; x < w; x++ {
			row += int(g.Pix[y*g.Stride+x])
			sum[(y+1)*(w+1)+x+1] = sum[y*(w+1)+x+1] + row
		}
	}

	out := image.NewGray(g.Bounds())
	r := lightTextWindow / 2
	for y := 0; y < h; y++ {
		y0, y1 := max(y-r, 0), min(y+r+1, h)
		for x := 0; x < w; x++ {
			x0, x1 := max(x-r, 0), min(x+r+1, w)
			total := sum[y1*(w+1)+x1] - sum[y0*(w+1)+x1] - sum[y1*(w+1)+x0] + sum[y0*(w+1)+x0]
			mean := total / ((x1 - x0) * (y1 - y0))
			out.Pix[y*out.Stride+x] = 255
			if mean < lightTextMaxBackground && int(g.Pix[y*g.Stride+x]) > mean+lightTextContrast {
				out.Pix[y*out.Stride+x] = 0
			}
		}
	}
	return out
}

func ocr(img image.Image, workDir string) ([]Word, error) {
	path := filepath.Join(workDir, "ocr.png")
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, err
	}

	// --psm 11: sparse text, UIs have no paragraphs.
	out, err := exec.Command("tesseract", path, "stdout", "--psm", "11", "tsv").Output()
	if err != nil {
		return nil, fmt.Errorf("running tesseract: %w", err)
	}
	return parseTSV(out), nil
}

func parseTSV(out []byte) []Word {
	var words []Word
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		f := strings.Split(scanner.Text(), "\t")
		if len(f) < 12 || f[0] != "5" { // 5 = word level
			continue
		}
		conf, _ := strconv.ParseFloat(f[10], 64)
		text := strings.TrimSpace(f[11])
		if text == "" || conf < minWordConfidence {
			continue
		}
		n := make([]int, 4)
		for i := range n {
			n[i], _ = strconv.Atoi(f[6+i])
		}
		words = append(words, Word{
			Text: text,
			Box:  image.Rect(n[0], n[1], n[0]+n[2], n[1]+n[3]),
			line: strings.Join(f[1:5], "/"),
		})
	}
	return words
}

func joinWords(words []Word) string {
	texts := make([]string, len(words))
	for i, w := range words {
		texts[i] = w.Text
	}
	return strings.Join(texts, " ")
}

func FindText(passes [][]Word, phrase string) (image.Rectangle, bool) {
	want := textKey(phrase)
	if want == "" {
		return image.Rectangle{}, false
	}
	var partial []image.Rectangle
	for _, words := range passes {
		for i, first := range words {
			if textKey(first.Text) == "" {
				continue // don't start a match at stray punctuation
			}
			box, end, ok := matchAt(words, i, want)
			if !ok {
				continue
			}
			if isWholeLine(words, i, end) {
				return box, true
			}
			partial = append(partial, box)
		}
	}
	if len(partial) > 0 {
		return partial[0], true
	}
	return image.Rectangle{}, false
}

func matchAt(words []Word, i int, want string) (image.Rectangle, int, bool) {
	var got string
	var box image.Rectangle
	for j := i; j < len(words) && words[j].line == words[i].line; j++ {
		got += textKey(words[j].Text)
		box = box.Union(words[j].Box)
		if got == want {
			return box, j + 1, true
		}
		if !strings.HasPrefix(want, got) {
			break
		}
	}
	return image.Rectangle{}, 0, false
}

func isWholeLine(words []Word, start, end int) bool {
	line := words[start].line
	for j, w := range words {
		if (j < start || j >= end) && w.line == line && textKey(w.Text) != "" {
			return false
		}
	}
	return true
}

func textKey(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, s)
}

func (m *Machine) WaitForText(t *testing.T, region image.Rectangle, phrase string, timeout time.Duration) image.Rectangle {
	t.Helper()
	box, ok := m.waitForText(t, region, phrase, timeout)
	if !ok {
		t.Fatalf("%q was not shown in %v within %s, last OCR:\n%s", phrase, region, timeout, m.lastOCR)
	}
	return box
}

func (m *Machine) TextShown(t *testing.T, region image.Rectangle, phrase string, timeout time.Duration) (image.Rectangle, bool) {
	t.Helper()
	return m.waitForText(t, region, phrase, timeout)
}

func (m *Machine) waitForText(t *testing.T, region image.Rectangle, phrase string, timeout time.Duration) (image.Rectangle, bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		if box, ok := FindText(m.ReadText(t, m.Screenshot(t), region), phrase); ok {
			return box, true
		}
		if time.Now().After(deadline) {
			return image.Rectangle{}, false
		}
		time.Sleep(pollInterval)
	}
}

func (m *Machine) ScreenSize(t *testing.T) image.Rectangle {
	t.Helper()
	if m.screen.Empty() {
		m.Screenshot(t)
	}
	return m.screen
}

func Center(r image.Rectangle) image.Point {
	return image.Pt((r.Min.X+r.Max.X)/2, (r.Min.Y+r.Max.Y)/2)
}
