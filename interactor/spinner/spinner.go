package spinner

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-runewidth"
)

type Spinner struct {
	wg     *sync.WaitGroup
	cancel context.CancelFunc
}

func (s Spinner) Stop() {
	s.cancel()
	s.wg.Wait()
}

func Spin(ctx context.Context, tips string, out *os.File, withDone bool) (spinner Spinner) {
	tips, ok := strings.CutSuffix(tips, "\n")
	spinnerChars := []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}
	ctx, spinner.cancel = context.WithCancel(ctx)
	spinner.wg = &sync.WaitGroup{}
	spinner.wg.Add(1)
	go func() {
		defer spinner.wg.Done()
		fmt.Fprintf(out, "%s %s", string(spinnerChars[0]), tips)
		for i := 1; ; i++ {
			spinnerChar := string(spinnerChars[i%len(spinnerChars)])
			line := fmt.Sprintf("%s %s", spinnerChar, tips)
			select {
			case <-ctx.Done():
				for range displayWidth(line) {
					fmt.Fprint(out, "\b")
				}
				if withDone {
					fmt.Fprintf(out, "\033[32m●\033[0m %s", tips)
				} else {
					fmt.Fprintf(out, "%s  \b\b", tips)
				}
				if ok {
					fmt.Fprint(out, "\n")
				}
				return
			default:
				time.Sleep(100 * time.Millisecond)
				for range displayWidth(line) {
					fmt.Fprint(out, "\b")
				}
				fmt.Fprint(out, line)
			}
		}
	}()
	return
}

func displayWidth(s string) int {
	width := 0
	for _, r := range s {
		width += runewidth.RuneWidth(r)
	}
	return width
}
