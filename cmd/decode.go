package cmd

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(decodeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// decodeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// decodeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// decodeCmd represents the decode command
var decodeCmd = &cobra.Command{
	Use:   "decode",
	Short: "A brief description of your command",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hexStr := args[0]

		var data []byte
		_, err := fmt.Sscanf(hexStr, "%x", &data)
		cobra.CheckErr(err)

		decoded, _, err := decodeMessagePack(data)
		cobra.CheckErr(err)
		fmt.Printf("%+v\n", decoded)
	},
}

func decodeMessagePack(data []byte) (interface{}, []byte, error) {
	if len(data) == 0 {
		return nil, nil, fmt.Errorf("empty MessagePack data")
	}

	code := data[0]
	data = data[1:]

	switch {
	case code <= 0x7f:
		// Positive FixInt
		return int(code), data, nil

	case code >= 0xe0:
		// Negative FixInt
		return int(int8(code)), data, nil

	case code >= 0x80 && code <= 0x8f:
		// FixMap
		length := int(code - 0x80)
		return decodeMap(data, length)

	case code >= 0x90 && code <= 0x9f:
		// FixArray
		length := int(code - 0x90)
		return decodeArray(data, length)

	case code >= 0xa0 && code <= 0xbf:
		// FixString
		length := int(code - 0xa0)
		strData := data[:length]
		data = data[length:]
		return string(strData), data, nil

	case code == 0xc0:
		// Nil
		return nil, data, nil

	case code == 0xc2:
		// False
		return false, data, nil

	case code == 0xc3:
		// True
		return true, data, nil

	case code == 0xca:
		// Float 32
		if len(data) < 4 {
			return nil, nil, fmt.Errorf("insufficient bytes for float32")
		}
		bits := binary.BigEndian.Uint32(data[:4])
		data = data[4:]
		return math.Float32frombits(bits), data, nil

	case code == 0xcb:
		// Float 64
		if len(data) < 8 {
			return nil, nil, fmt.Errorf("insufficient bytes for float64")
		}
		bits := binary.BigEndian.Uint64(data[:8])
		data = data[8:]
		return math.Float64frombits(bits), data, nil

	case code == 0xcc:
		// Unsigned 8-bit integer
		if len(data) < 1 {
			return nil, nil, fmt.Errorf("insufficient bytes for uint8")
		}
		return uint(data[0]), data[1:], nil

	case code == 0xcd:
		// Unsigned 16-bit integer
		if len(data) < 2 {
			return nil, nil, fmt.Errorf("insufficient bytes for uint16")
		}
		value := binary.BigEndian.Uint16(data[:2])
		data = data[2:]
		return uint(value), data, nil

	case code == 0xce:
		// Unsigned 32-bit integer
		if len(data) < 4 {
			return nil, nil, fmt.Errorf("insufficient bytes for uint32")
		}
		value := binary.BigEndian.Uint32(data[:4])
		data = data[4:]
		return uint(value), data, nil

	case code == 0xcf:
		// Unsigned 64-bit integer
		if len(data) < 8 {
			return nil, nil, fmt.Errorf("insufficient bytes for uint64")
		}
		value := binary.BigEndian.Uint64(data[:8])
		data = data[8:]
		return uint64(value), data, nil

	case code == 0xd0:
		// Signed 8-bit integer
		if len(data) < 1 {
			return nil, nil, fmt.Errorf("insufficient bytes for int8")
		}
		return int(int8(data[0])), data[1:], nil

	case code == 0xd1:
		// Signed 16-bit integer
		if len(data) < 2 {
			return nil, nil, fmt.Errorf("insufficient bytes for int16")
		}
		value := int16(binary.BigEndian.Uint16(data[:2]))
		data = data[2:]
		return int(value), data, nil

	case code == 0xd2:
		// Signed 32-bit integer
		if len(data) < 4 {
			return nil, nil, fmt.Errorf("insufficient bytes for int32")
		}
		value := int32(binary.BigEndian.Uint32(data[:4]))
		data = data[4:]
		return int(value), data, nil

	case code == 0xd3:
		// Signed 64-bit integer
		if len(data) < 8 {
			return nil, nil, fmt.Errorf("insufficient bytes for int64")
		}
		value := int64(binary.BigEndian.Uint64(data[:8]))
		data = data[8:]
		return value, data, nil

	case code == 0xd9:
		// String 8
		if len(data) < 1 {
			return nil, nil, fmt.Errorf("insufficient bytes for string 8")
		}
		length := int(data[0])
		data = data[1:]
		return decodeString(data, length)

	case code == 0xda:
		// String 16
		if len(data) < 2 {
			return nil, nil, fmt.Errorf("insufficient bytes for string 16")
		}
		length := int(binary.BigEndian.Uint16(data[:2]))
		data = data[2:]
		return decodeString(data, length)

	case code == 0xdb:
		// String 32
		if len(data) < 4 {
			return nil, nil, fmt.Errorf("insufficient bytes for string 32")
		}
		length := int(binary.BigEndian.Uint32(data[:4]))
		data = data[4:]
		return decodeString(data, length)

	case code == 0xdc:
		// Array 16
		if len(data) < 2 {
			return nil, nil, fmt.Errorf("insufficient bytes for array 16")
		}
		length := int(binary.BigEndian.Uint16(data[:2]))
		data = data[2:]
		return decodeArray(data, length)

	case code == 0xdd:
		// Array 32
		if len(data) < 4 {
			return nil, nil, fmt.Errorf("insufficient bytes for array 32")
		}
		length := int(binary.BigEndian.Uint32(data[:4]))
		data = data[4:]
		return decodeArray(data, length)

	case code == 0xde:
		// Map 16
		if len(data) < 2 {
			return nil, nil, fmt.Errorf("insufficient bytes for map 16")
		}
		length := int(binary.BigEndian.Uint16(data[:2]))
		data = data[2:]
		return decodeMap(data, length)

	case code == 0xdf:
		// Map 32
		if len(data) < 4 {
			return nil, nil, fmt.Errorf("insufficient bytes for map 32")
		}
		length := int(binary.BigEndian.Uint32(data[:4]))
		data = data[4:]
		return decodeMap(data, length)

	default:
		return nil, nil, fmt.Errorf("unsupported MessagePack data type: %#x", code)
	}
}

func decodeArray(data []byte, length int) ([]interface{}, []byte, error) {
	var array []interface{}
	var err error

	for i := 0; i < length; i++ {
		var element interface{}
		element, data, err = decodeMessagePack(data)
		if err != nil {
			return nil, nil, err
		}
		array = append(array, element)
	}

	return array, data, nil
}

func decodeMap(data []byte, length int) (map[string]interface{}, []byte, error) {
	m := make(map[string]interface{})
	var err error

	for i := 0; i < length; i++ {
		var key interface{}
		key, data, err = decodeMessagePack(data)
		if err != nil {
			return nil, nil, err
		}

		keyStr, ok := key.(string)
		if !ok {
			return nil, nil, fmt.Errorf("map key is not a string: %v", key)
		}

		var value interface{}
		value, data, err = decodeMessagePack(data)
		if err != nil {
			return nil, nil, err
		}

		m[keyStr] = value
	}

	return m, data, nil
}

func decodeString(data []byte, length int) (string, []byte, error) {
	if len(data) < length {
		return "", nil, fmt.Errorf("insufficient bytes for string")
	}

	str := string(data[:length])
	data = data[length:]
	return str, data, nil
}
