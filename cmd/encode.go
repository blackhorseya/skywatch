package cmd

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// encodeCmd represents the encode command
var encodeCmd = &cobra.Command{
	Use:   "encode",
	Short: "A brief description of your command",
}

func init() {
	encodeCmd.AddCommand(encodeJsonCmd)

	rootCmd.AddCommand(encodeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// encodeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// encodeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

var encodeJsonCmd = &cobra.Command{
	Use:   "json",
	Short: "Encode JSON",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]

		var data map[string]interface{}
		err := json.Unmarshal([]byte(input), &data)
		cobra.CheckErr(err)

		encoded, err := encodeMessagePack(data)
		cobra.CheckErr(err)

		fmt.Printf("%x\n", encoded)
	},
}

func encodeMessagePack(data map[string]interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)

	// write map header
	err := binary.Write(buf, binary.BigEndian, 0x80|byte(len(data)))
	if err != nil {
		return nil, err
	}

	for key, value := range data {
		// write key
		writeString(buf, key)

		// write value
		writeValue(buf, value)
	}

	return buf.Bytes(), nil
}

func writeString(buf *bytes.Buffer, str string) {
	// write string length
	binary.Write(buf, binary.BigEndian, byte(len(str)))

	// write string
	buf.WriteString(str)
}

func writeValue(buf *bytes.Buffer, value interface{}) {
	switch v := value.(type) {
	case int:
		// write int value
		binary.Write(buf, binary.BigEndian, int32(v))
	case float64:
		// Write float64 value
		binary.Write(buf, binary.BigEndian, v)
	case string:
		// write string value
		writeString(buf, v)
	case bool:
		// Write bool value
		if v {
			buf.Write([]byte{0xc3}) // True
		} else {
			buf.Write([]byte{0xc2}) // False
		}
	case nil:
		// Write nil value
		buf.Write([]byte{0xc0}) // Nil
	default:
		// unsupported type
		fmt.Printf("unsupported type: %T\n", v)
	}
}
