package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/text/encoding/simplifiedchinese"
)

func main() {
	cobra.CheckErr(rootCmd.Execute())
}

var (
	version = "1.0"

	UTF8     = false
	GB18030  = false
	GBK      = false
	HZGB2312 = false

	rootCmd = &cobra.Command{
		Use:   "decoder",
		Short: "",
		Long:  " \r\n Read file format character tool!",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	readCmd = &cobra.Command{
		Use:   "read",
		Short: "Read file with specify path",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				log.Fatalf("Invalid args, nest error: %v\r\n", args)
			}
			file, err := os.OpenFile(args[0], os.O_RDONLY, 0644)
			if err != nil {
				log.Printf("[Fatal] Open file failure, nest error: %v\r\n", err)
				os.Exit(-1)
			}
			defer file.Close()

			var (
				reader  = bufio.NewReader(file)
				buf     = make([]byte, 8*1024)
				charset = "utf-8"
			)
			switch {
			case GBK:
				charset = "gbk"
			case HZGB2312:
				charset = "hzgb2312"
			case GB18030:
				charset = "gb18030"
			default:
				charset = "utf-8"
			}

			for {
				n, err := reader.Read(buf)
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Printf("[Fatal] Read file failure, nest error: %v\r\n", err)
					os.Exit(-1)
				}
				fmt.Print(BytesToString(charset, buf[:n]))
			}
			fmt.Println()
		},
	}
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version about decoder",
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

	rootCmd.AddCommand(versionCmd)

	readCmd.Flags().BoolVarP(&GBK, "gbk", "", false, "gbk mode")
	readCmd.Flags().BoolVarP(&UTF8, "utf8", "", false, "utf-8 mode")
	readCmd.Flags().BoolVarP(&GB18030, "gb18030", "", false, "gb18030 mode")
	readCmd.Flags().BoolVarP(&HZGB2312, "hzgb2312", "", false, "hzgb2312 mode")
	rootCmd.AddCommand(readCmd)
}

// BytesToString 字节转换为字符串
func BytesToString(charset string, buf []byte) string {
	var str string
	switch charset {
	case "gb18030":
		tmp, _ := simplifiedchinese.GB18030.NewDecoder().Bytes(buf)
		str = string(tmp)
	case "gbk":
		tmp, _ := simplifiedchinese.GBK.NewDecoder().Bytes(buf)
		str = string(tmp)
	case "hzgb2312":
		tmp, _ := simplifiedchinese.HZGB2312.NewDecoder().Bytes(buf)
		str = string(tmp)
	case "utf-8":
		fallthrough
	default:
		str = string(buf)
	}
	return str
}

func printClientVersion() {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("decoder version (Current): %s", version))
	fmt.Println(buf.String())

}
