package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/term"
)

type UserEventKind int

type UserEvent struct {
	Kind UserEventKind
	Err  error
}

const (
	UserEventKindUp UserEventKind = iota
	UserEventKindDown
	UserEventKindLeft
	UserEventKindRight
	UserEventKindError
	UserEventKindQuit
)

func readInput(ctx context.Context, r io.Reader, events chan<- UserEvent) {
	defer close(events)
	br := bufio.NewReader(r)
	var evt UserEvent
	for {
		ch, _, err := br.ReadRune()
		if err != nil {
			evt = UserEvent{UserEventKindError, err}
		} else {
			switch ch {
			case 'q':
				evt = UserEvent{UserEventKindQuit, nil}
			case '\x1b':
				next, _, err := br.ReadRune()
				if err != nil {
					evt = UserEvent{UserEventKindError, err}
				}
				if next != '[' {
					continue
				}

				next, _, err = br.ReadRune()
				if err != nil {
					evt = UserEvent{UserEventKindError, err}
				}
				switch next {
				case 'A':
					evt = UserEvent{UserEventKindUp, nil}
				case 'B':
					evt = UserEvent{UserEventKindDown, nil}
				case 'C':
					evt = UserEvent{UserEventKindRight, nil}
				case 'D':
					evt = UserEvent{UserEventKindLeft, nil}
				default:
					continue
				}
			default:
				continue
			}
		}

		select {
		case events <- evt:
		case <-ctx.Done():
			return
		}
	}
}

func run() error {
	old, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	defer func(fd int, oldState *term.State) {
		_ = term.Restore(fd, oldState)
	}(int(os.Stdin.Fd()), old)

	w, h, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}

	x, y := w/2, h/2

	events := make(chan UserEvent)
	defer close(events)

	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go readInput(ctx, os.Stdin, events)

	for {
		select {
		case evt := <-events:
			switch evt.Kind {
			case UserEventKindError:
				return evt.Err
			case UserEventKindUp:
				y--
			case UserEventKindDown:
				y++
			case UserEventKindLeft:
				x--
			case UserEventKindRight:
				x++
			case UserEventKindQuit:
				fmt.Print("\x1b[2J")
				return nil
			}
		case <-ticker.C:
			fmt.Print("\x1b[2J")
			fmt.Printf("\x1b[%d;%dH@", y, x)
		}
	}
}

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
