package cmd

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/spf13/cobra"
)

const (
	mathMinInt32  = math.MinInt32
	mathMaxInt32  = math.MaxInt32
	mathMaxUint32 = math.MaxUint32
	mathMaxUint8  = math.MaxUint8
	mathMaxUint16 = math.MaxUint16
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

		var data interface{}
		err := json.Unmarshal([]byte(input), &data)
		cobra.CheckErr(err)

		encoded, err := convertToMsgpack(data)
		cobra.CheckErr(err)
		fmt.Printf("%x\n", encoded)
	},
}

func convertToMsgpack(data interface{}) ([]byte, error) {
	var msgpackData []byte
	var err error

	// Use reflection to dynamically handle different data types.
	switch v := data.(type) {
	case nil:
		msgpackData = []byte{0xc0} // nil
	case bool:
		if v {
			msgpackData = []byte{0xc3} // true
		} else {
			msgpackData = []byte{0xc2} // false
		}
	case int:
		if v >= 0 && v <= 127 {
			msgpackData = []byte{byte(v)} // positive fixint
		} else if v >= -32 && v <= -1 {
			msgpackData = []byte{byte(0xe0 | (v & 0x1f))} // negative fixint
		} else if v >= mathMinInt32 && v <= mathMaxInt32 {
			msgpackData = []byte{0xd2}
			uint32Data := uint32(v)
			buffer := make([]byte, 4)
			binary.BigEndian.PutUint32(buffer, uint32Data)
			msgpackData = append(msgpackData, buffer...)
		} else {
			return nil, fmt.Errorf("integer value out of MessagePack range")
		}
	case int64:
		msgpackData = []byte{0xd3}
		uint64Data := uint64(v)
		buffer := make([]byte, 8)
		binary.BigEndian.PutUint64(buffer, uint64Data)
		msgpackData = append(msgpackData, buffer...)
	case uint:
		if v <= mathMaxUint32 {
			msgpackData = []byte{0xcc}
			buffer := make([]byte, 4)
			binary.BigEndian.PutUint32(buffer, uint32(v))
			msgpackData = append(msgpackData, buffer...)
		} else {
			return nil, fmt.Errorf("unsigned integer value out of MessagePack range")
		}
	case uint64:
		msgpackData = []byte{0xcf}
		buffer := make([]byte, 8)
		binary.BigEndian.PutUint64(buffer, v)
		msgpackData = append(msgpackData, buffer...)
	case float32:
		msgpackData = []byte{0xca}
		bits := math.Float32bits(v)
		buffer := make([]byte, 4)
		binary.BigEndian.PutUint32(buffer, bits)
		msgpackData = append(msgpackData, buffer...)
	case float64:
		msgpackData = []byte{0xcb}
		bits := math.Float64bits(v)
		buffer := make([]byte, 8)
		binary.BigEndian.PutUint64(buffer, bits)
		msgpackData = append(msgpackData, buffer...)
	case string:
		strBytes := []byte(v)
		length := len(strBytes)
		if length <= 31 {
			msgpackData = []byte{byte(0xa0 | length)}
		} else if length <= mathMaxUint8 {
			msgpackData = []byte{0xd9, byte(length)}
		} else if length <= mathMaxUint16 {
			msgpackData = []byte{0xda, byte(length >> 8), byte(length)}
		} else if length <= mathMaxUint32 {
			msgpackData = []byte{0xdb, byte(length >> 24), byte(length >> 16), byte(length >> 8), byte(length)}
		} else {
			return nil, fmt.Errorf("string length exceeds MessagePack maximum")
		}
		msgpackData = append(msgpackData, strBytes...)
	case []interface{}:
		arrayLength := len(v)
		if arrayLength <= 15 {
			msgpackData = []byte{byte(0x90 | arrayLength)}
		} else if arrayLength <= mathMaxUint16 {
			msgpackData = []byte{0xdc, byte(arrayLength >> 8), byte(arrayLength)}
		} else if arrayLength <= mathMaxUint32 {
			msgpackData = []byte{0xdd, byte(arrayLength >> 24), byte(arrayLength >> 16), byte(arrayLength >> 8), byte(arrayLength)}
		} else {
			return nil, fmt.Errorf("array length exceeds MessagePack maximum")
		}
		for _, item := range v {
			itemData, err := convertToMsgpack(item)
			if err != nil {
				return nil, err
			}
			msgpackData = append(msgpackData, itemData...)
		}
	case map[string]interface{}:
		mapLength := len(v)
		if mapLength <= 15 {
			msgpackData = []byte{byte(0x80 | mapLength)}
		} else if mapLength <= mathMaxUint16 {
			msgpackData = []byte{0xde, byte(mapLength >> 8), byte(mapLength)}
		} else if mapLength <= mathMaxUint32 {
			msgpackData = []byte{0xdf, byte(mapLength >> 24), byte(mapLength >> 16), byte(mapLength >> 8), byte(mapLength)}
		} else {
			return nil, fmt.Errorf("map length exceeds MessagePack maximum")
		}
		for key, value := range v {
			keyData, err := convertToMsgpack(key)
			if err != nil {
				return nil, err
			}
			valueData, err := convertToMsgpack(value)
			if err != nil {
				return nil, err
			}
			msgpackData = append(msgpackData, keyData...)
			msgpackData = append(msgpackData, valueData...)
		}
	case time.Time:
		// Serialize time.Time as a string.
		msgpackData, err = convertToMsgpack(v.Format(time.RFC3339))
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported data type: %v", reflect.TypeOf(data))
	}

	return msgpackData, nil
}
