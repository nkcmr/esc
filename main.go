package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

func runWithError(fn func(cmd *cobra.Command, pargs []string) error) func(*cobra.Command, []string) {
	return func(c *cobra.Command, s []string) {
		if err := fn(c, s); err != nil {
			fatal(err.Error())
			return
		}
	}
}

func xor(b ...bool) bool {
	result := false
	for _, value := range b {
		result = result != value
	}
	return result
}

func or(b ...bool) bool {
	result := false
	for _, value := range b {
		result = result || value
	}
	return result
}

const (
	_ = int(iota)
	ctxtUnquoted
	ctxtSingleQuotes
	ctxtDoubleQuotes
)

func escape(ctxt int, r io.Reader, w io.Writer) error {
	buf := bufio.NewReader(r)
	wbuf := bufio.NewWriter(w)
	for {
		b, err := buf.ReadByte()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}
		const singleQuote = '\''
		const doubleQuote = '"'
		const space = ' '
		const backslash = '\\'

		switch b {
		case space:
			if ctxt == ctxtUnquoted {
				if _, err := wbuf.Write([]byte{
					backslash,
					space,
				}); err != nil {
					return err
				}
				continue
			}
		case singleQuote:
			switch ctxt {
			case ctxtSingleQuotes, ctxtUnquoted:
				if _, err := wbuf.Write([]byte{
					backslash,
					singleQuote,
				}); err != nil {
					return err
				}
				continue
			}
		case doubleQuote:
			switch ctxt {
			case ctxtDoubleQuotes, ctxtUnquoted:
				if _, err := wbuf.Write([]byte{
					backslash,
					doubleQuote,
				}); err != nil {
					return err
				}
				continue
			}
		}

		if err := wbuf.WriteByte(b); err != nil {
			return err
		}
	}

	if err := wbuf.Flush(); err != nil {
		return err
	}

	return nil
}

func rootCommand() *cobra.Command {
	var cliargs struct {
		ctxtUnquoted     bool
		ctxtSingleQuotes bool
		ctxtDoubleQuotes bool
		perline          bool
	}
	cmd := &cobra.Command{
		Use: "esc",
		Run: runWithError(func(cmd *cobra.Command, pargs []string) error {
			if !xor(cliargs.ctxtUnquoted, cliargs.ctxtSingleQuotes, cliargs.ctxtDoubleQuotes) {
				return fmt.Errorf("must enable exactly 1 context")
			}
			var ctxt int
			switch {
			case cliargs.ctxtUnquoted:
				ctxt = ctxtUnquoted
			case cliargs.ctxtSingleQuotes:
				ctxt = ctxtSingleQuotes
			case cliargs.ctxtDoubleQuotes:
				ctxt = ctxtDoubleQuotes
			default:
				return fmt.Errorf("no context specified")
			}

			if cliargs.perline {
				stdout := bufio.NewWriter(os.Stdout)
				lineScan := bufio.NewScanner(os.Stdin)
				for lineScan.Scan() {
					if err := escape(ctxt, bytes.NewReader(lineScan.Bytes()), stdout); err != nil {
						return err
					}
					if err := stdout.WriteByte('\n'); err != nil {
						return err
					}
					if err := stdout.Flush(); err != nil {
						return err
					}
				}
				if err := lineScan.Err(); err != nil {
					return fmt.Errorf("failed to read stdin: %s", err.Error())
				}
			} else {
				if err := escape(ctxt, os.Stdin, os.Stdout); err != nil {
					return fmt.Errorf("failed to read stdin: %s", err.Error())
				}
			}

			return nil
		}),
	}
	cmd.Flags().BoolVarP(&cliargs.perline, "per-line", "l", true, "will escape each input line individually")
	cmd.Flags().BoolVarP(&cliargs.ctxtUnquoted, "unquoted", "u", false, "signifies that the strings should be escaped for an unquoted evaluation")
	cmd.Flags().BoolVarP(&cliargs.ctxtSingleQuotes, "single-quoted", "s", false, "signifies that the strings should be escaped for an single quoted evaluation")
	cmd.Flags().BoolVarP(&cliargs.ctxtDoubleQuotes, "double-quoted", "d", false, "signifies that the strings should be escaped for an double quoted evaluation")
	return cmd
}

func main() {
	_ = rootCommand().Execute()
}

func fatal(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "esc: error: "+format+"\n", a...)
	os.Exit(1)
}
