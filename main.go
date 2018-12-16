package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func mustRead(p []byte) int {
	n, err := os.Stdin.Read(p)
	if err == io.EOF {
		os.Exit(0)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "read: %s\n", err)
		os.Exit(1)
	}
	return n
}

func mustWrite(p []byte) {
	_, err := os.Stdout.Write(p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "write: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	var bufsize, skip, frame, offset, pass int
	flag.IntVar(&bufsize, "buffer", 1<<20, "buffer `size` in bytes")
	flag.IntVar(&skip, "skip", 0, "`bytes` to skip at start of input")
	flag.IntVar(&frame, "frame", 1, "frame `size` in bytes")
	flag.IntVar(&offset, "offset", 0, "`bytes` to skip at the beginning of each frame")
	flag.IntVar(&pass, "pass", 1, "`bytes` to output from each frame")
	flag.Parse()

	if frame < 1 {
		fmt.Fprintf(os.Stderr, "invalid argument; cannot use frame<1\n")
		os.Exit(2)
	} else if offset+pass > frame {
		fmt.Fprintf(os.Stderr, "invalid argument; cannot use offset+pass > frame\n")
		os.Exit(2)
	} else if bufsize < frame {
		fmt.Fprintf(os.Stderr, "invalid argument; cannot use bufsize < frame\n")
		os.Exit(2)
	}

	inbuf := make([]byte, bufsize)
	todo := mustRead(inbuf)

	for ; skip >= todo; todo, skip = mustRead(inbuf), skip-todo {
	}
	todo -= skip
	copy(inbuf[:todo], inbuf[skip:])

	out := make(chan []byte)
	go func() {
		for buf := range out {
			mustWrite(buf)
		}
	}()

	for ; ; todo += mustRead(inbuf[todo:]) {
		done := 0
		outbuf := make([]byte, 0, int(int64(todo)/int64(frame)*int64(pass)))
		for ; done+frame <= todo; done += frame {
			outbuf = append(outbuf, inbuf[done+offset:done+offset+pass]...)
		}
		out <- outbuf

		if done > 0 && done < todo {
			copy(inbuf, inbuf[done:todo])
		}
		todo -= done
	}
}
