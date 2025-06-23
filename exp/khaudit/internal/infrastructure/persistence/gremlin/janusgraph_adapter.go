package gremlin

import (
	"encoding/binary"
	"errors"
	"fmt"

	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

const (
	valueFlagNone byte = 0
)

func readBytes(data *[]byte, i *int, len int) []byte {
	tmp := (*data)[*i : *i+len]
	*i += len
	return tmp
}

func readUint32Safe(data *[]byte, i *int) uint32 {
	return binary.BigEndian.Uint32(readBytes(data, i, 4))
}

func readByteSafe(data *[]byte, i *int) byte {
	*i++
	return (*data)[*i-1]
}

func readLongSafe(data *[]byte, i *int) int64 {
	return int64(binary.BigEndian.Uint64(readBytes(data, i, 8)))
}

func readString(data *[]byte, i *int) (interface{}, error) {
	sz := int(readUint32Safe(data, i))
	if sz == 0 {
		return "", nil
	}
	*i += sz
	return string((*data)[*i-sz : *i]), nil
}

// janusgraphRelationIdentifierReader reads a JanusGraph relation identifier from the given byte slice.
// It expects the data to be in a specific format and returns a map containing the relation identifier details.
//
// The format is as follows:
// - 4 bytes: custom data type (0x1001)
// - 1 byte: value flag (0x00)
// - 1 byte: outVertexId type (0x00 for long, 0x01 for string)
//   - if 0x00: 8 bytes: outVertexId (long)
//   - if 0x01: 4 bytes: outVertexId length, followed by outVertexId string
//
// - 8 bytes: typeId (long)
// - 8 bytes: relationId (long)
// - 1 byte: inVertexId type (0x00 for long, 0x01 for string)
//   - if 0x00: 8 bytes: inVertexId (long)
//   - if 0x01: 4 bytes: inVertexId length, followed by inVertexId string
func janusgraphRelationIdentifierReader(data *[]byte, i *int) (any, error) {
	const relationIdentifierType = 0x1001
	const (
		longMarker   = 0
		stringMarker = 1
	)

	// check custom data type
	customDataType := readUint32Safe(data, i)
	if customDataType != relationIdentifierType {
		return nil, fmt.Errorf("unknown type code. got 0x%x, expected 0x%x", customDataType, relationIdentifierType)
	}

	// value flag, expect this to be non-nullable
	if readByteSafe(data, i) != valueFlagNone {
		return nil, errors.New("expected non-null value")
	}

	// outVertexId
	var (
		outVertexId any
		err         error
	)
	if readByteSafe(data, i) == stringMarker {
		outVertexId, err = readString(data, i)
		if err != nil {
			return nil, fmt.Errorf("unable to read outVertexId: %w", err)
		}
	} else {
		outVertexId = readLongSafe(data, i)
	}

	// typeId
	typeId := readLongSafe(data, i)
	// relationId
	relationId := readLongSafe(data, i)

	// inVertexId
	var inVertexId any
	if readByteSafe(data, i) == stringMarker {
		inVertexId, err = readString(data, i)
		if err != nil {
			return nil, fmt.Errorf("unable to read inVertexId: %w", err)
		}
	} else {
		inVertexId = readLongSafe(data, i)
	}

	return map[string]any{
		"outVertexId": outVertexId,
		"typeId":      typeId,
		"relationId":  relationId,
		"inVertexId":  inVertexId,
	}, nil
}

func init() {
	// https://issues.apache.org/jira/browse/TINKERPOP-2802
	gremlingo.RegisterCustomTypeReader("janusgraph.RelationIdentifier", janusgraphRelationIdentifierReader)
}
