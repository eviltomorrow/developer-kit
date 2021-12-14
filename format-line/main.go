package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func main() {
	cobra.CheckErr(rootCmd.Execute())
}

var (
	version = "1.0"
	path    string
	rootCmd = &cobra.Command{
		Use:   "fomart-line",
		Short: "",
		Long:  " \r\n Read file format character tool!",
		Run: func(cmd *cobra.Command, args []string) {
			reader, cancel, err := getReader(path)
			if err != nil {
				log.Fatalf("panic: get reader failure, nest error: %v\r\n", err)
			}
			defer cancel()

			printStdout(reader)
		},
	}

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version about format-line",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			printClientVersion()
		},
	}
)

func init() {
	rootCmd.CompletionOptions = cobra.CompletionOptions{
		DisableDefaultCmd: true,
	}
	rootCmd.Flags().StringVarP(&path, "path", "p", "", "path of file")
	rootCmd.AddCommand(versionCmd)
}

func printClientVersion() {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("format version (Current): %s", version))
	fmt.Println(buf.String())
}

func printStdout(r io.Reader) {
	var (
		reader = bufio.NewReader(r)
		buf    bytes.Buffer
		dirty  bytes.Buffer

		mark       bool
		notsupport = make([]string, 0, 16)
	)
	for {
		b, err := reader.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("panic: read byte from stdin failure, nest error: %v\r\n", err)
		}
		if b == '\\' && !mark {
			mark = true
			continue
		}

		if mark {
			switch b {
			case 'n':
				if dirty.Len() != 0 {
					fmt.Print(dirty.String())
					dirty.Reset()
				}
				fmt.Println(buf.String())
				buf.Reset()

			case '\\':
				buf.WriteByte('\\')
				dirty.Reset()

			case 'r':
				dirty.Reset()
				if _, err := dirty.Write(buf.Bytes()); err != nil {
					log.Fatalf("panic: dirty write failure, nest error: %v\r\n", err)
				}
				buf.Reset()

			case '"':
				buf.WriteByte('"')
				dirty.Reset()

			case '\'':
				buf.WriteByte('\'')
				dirty.Reset()

			case 't':
				for i := 0; i < 8; i++ {
					buf.WriteByte(' ')
				}
				dirty.Reset()

			case '?':
				buf.WriteByte('?')
				dirty.Reset()

			case 'b':
				n := buf.Len() - 1
				if n >= 0 {
					buf.Truncate(n)
				}

				dirty.Reset()

			default:
				notsupport = append(notsupport, fmt.Sprintf("\\%c", b))
			}
		} else {
			buf.WriteByte(b)
		}
		mark = false
	}
	if buf.Len() != 0 {
		fmt.Printf("%s\r\n", buf.String())
	}
	if len(notsupport) != 0 {
		red := color.New(color.FgRed).SprintFunc()
		log.Printf("[%s]=>Exist not support escape character: <%s>\r\n", red("Fatal"), strings.Join(notsupport, ","))
	}
}

func getReader(path string) (io.Reader, func(), error) {
	if path != "" {
		file, err := os.OpenFile(path, os.O_RDONLY, 0644)
		if err != nil {
			return nil, nil, err
		}
		return file, func() { file.Close() }, nil
	} else {
		stat, err := os.Stdin.Stat()
		if err != nil {
			log.Fatalf("stdin stat failure, nest error: %v\r\n", err)
		}
		if stat == nil {
			log.Fatalf("panic: stdin stat is nil\r\n")
		}
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			log.Fatalf("panic: invalid stdin stat mode[%d]\r\n", (stat.Mode() & os.ModeCharDevice))
		}
		return os.Stdin, func() {}, nil
	}
}
