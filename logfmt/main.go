package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
)

const (
	KB = 1024
	MB = 1024 * 1024
)

func main() {
	if len(os.Args) == 1 {
		log.Fatal("Please set log text file. eg. ./logfmt [FILE]")
	}

	var files = os.Args[1:]
	for _, path := range files {
		if err := readFile(path); err != nil {
			log.Fatal(err)
		}
	}
}

func readFile(path string) error {
	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	var buf [32 * KB]byte
	for {
		n, err := file.Read(buf[0:])
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		var data = parseEscapeCharacter(buf[:n])
		fmt.Print(data)

	}
	return nil
}

func parseEscapeCharacter(buf []byte) string {
	var buffer bytes.Buffer
	for i := 0; i < len(buf); i++ {
		var b = buf[i]
		if b == '\\' && i < len(buf)-1 {
			var b1 = buf[i+1]
			switch b1 {
			case '\\':
				if i < len(buf)-2 {
					var b2 = buf[i+2]
					if b2 != '\\' {
						continue
					} else {
						buffer.WriteByte('\\')
						i += 2
					}
				}
				continue
			case 'r':
				buffer.WriteByte('\r')
				i++
			case 't':
				buffer.WriteByte('\t')
				i++
			case 'n':
				buffer.WriteByte('\n')
				i++
			default:
			}
		} else {
			buffer.WriteByte(b)
		}

	}
	return buffer.String()
}
