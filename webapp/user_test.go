package webapp

import (
	"bufio"
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

const userDataURL = `user=%7B%22id%22:123456789%2C%22first_name%22:%22Val%C3%A9ry%22%2C%22last_name%22:%22Zhmishenk%C3%B6%22%2C%22username%22:%22zhmih17%22%2C%22language_code%22:%22ru%22%2C%22is_premium%22:true%2C%22allows_write_to_pm%22:true%7D`
const userDataHash = `user={"id":123456789,"first_name":"Valéry","last_name":"Zhmishenkö","username":"zhmih17","language_code":"ru","is_premium":true,"allows_write_to_pm":true}`

func Test_DecodeUser(t *testing.T) {
	buf := bytes.NewBuffer([]byte(userDataURL))
	r := bufio.NewReaderSize(buf, buf.Len())
	wantedResults := InitDataList{
		&User{
			ID:              123456789,
			FirstName:       "Valéry",     // %C3%A9
			LastName:        "Zhmishenkö", // %C3%B6
			Username:        "zhmih17",
			LanguageCode:    "ru",
			IsPremium:       true,
			AllowsWriteToPM: true,
		},
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

func Test_SerializeUser(t *testing.T) {
	expect1 := []byte(userDataURL)
	expect2 := []byte(userDataHash)
	results := InitDataList{
		&User{
			ID:              123456789,
			FirstName:       "Valéry",     // %C3%A9
			LastName:        "Zhmishenkö", // %C3%B6
			Username:        "zhmih17",
			LanguageCode:    "ru",
			IsPremium:       true,
			AllowsWriteToPM: true,
		},
	}

	b1, err := results.Serialize('&')
	assert.Nil(t, err)
	assert.Equal(t, expect1, b1)

	b2, err := results.Serialize('\n')
	assert.Nil(t, err)
	assert.Equal(t, expect2, b2)
}
