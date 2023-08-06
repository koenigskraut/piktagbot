package webapp

import (
	"bufio"
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

const simpleFieldsURL = `query_id=123&auth_date=12345678&prefix=query&hash=abcdef123456789`
const simpleFieldsHash = `query_id=123
auth_date=12345678
prefix=query
hash=abcdef123456789`

func Test_DecodeSimpleFields(t *testing.T) {
	buf := bytes.NewBuffer([]byte(simpleFieldsURL))
	r := bufio.NewReaderSize(buf, buf.Len())
	wantedResults := InitData{
		&QueryID{Data: "123"},
		&AuthDate{Data: 12345678},
		&Prefix{Data: "query"},
		&Hash{Data: "abcdef123456789"},
	}
	i := 0
	for {
		field, err := DecodeField(r)
		if err != nil {
			assert.ErrorIs(t, err, io.EOF)
			break
		}
		assert.Nil(t, err)
		assert.EqualValues(t, wantedResults[i], field)
		i++
	}
}

func Test_SerializeSimpleFields(t *testing.T) {
	expect1 := []byte(simpleFieldsURL)
	expect2 := []byte(simpleFieldsHash)
	results := InitData{
		&QueryID{Data: "123"},
		&AuthDate{Data: 12345678},
		&Prefix{Data: "query"},
		&Hash{Data: "abcdef123456789"},
	}
	b1, err := results.Serialize('&')
	assert.Nil(t, err)
	assert.Equal(t, expect1, b1)
	b2, err := results.Serialize('\n')
	assert.Nil(t, err)
	assert.Equal(t, expect2, b2)
}
