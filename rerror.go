package tp

import (
	"encoding/json"
	"strconv"
	"unsafe"

	"github.com/tidwall/gjson"

	"github.com/henrylee2cn/goutil"
	"github.com/henrylee2cn/teleport/utils"
)

type (
	// Rerror error only for reply message
	Rerror struct {
		// Code error code
		ErrorCode int32
		// Message the error message displayed to the user (optional)
		ErrorMsg string
		// Content
		Data Content
	}
	// Content
	Content struct {
		// Reason the cause of the error for debugging (optional)
		Reason string
	}

)

var (
	_ json.Marshaler   = new(Rerror)
	_ json.Unmarshaler = new(Rerror)

	reA = []byte(`{"error_code":`)
	reB = []byte(`,"error_msg":`)
	reC = []byte(`,"data":{reason:`)
	reD = []byte(`,"data":{}`)
)

// NewRerror creates a *Rerror.
func NewRerror(code int32, message, reason string) *Rerror {
	return &Rerror{
		ErrorCode:    code,
		ErrorMsg: message,
		Data:  Content{Reason: reason},
	}
}

// NewRerrorFromMeta creates a *Rerror from 'X-Reply-Error' metadata.
// Return nil if there is no 'X-Reply-Error' in metadata.
func NewRerrorFromMeta(meta *utils.Args) *Rerror {
	b := meta.Peek(MetaRerror)
	if len(b) == 0 {
		return nil
	}
	r := new(Rerror)
	r.UnmarshalJSON(b)
	return r
}

// SetToMeta sets self to 'X-Reply-Error' metadata.
func (r *Rerror) SetToMeta(meta *utils.Args) {
	b, _ := r.MarshalJSON()
	if len(b) == 0 {
		return
	}
	meta.Set(MetaRerror, goutil.BytesToString(b))
}

// Copy returns the copy of Rerror
func (r Rerror) Copy() *Rerror {
	return &r
}

// SetMessage sets the error message displayed to the user.
func (r *Rerror) SetMessage(message string) *Rerror {
	r.ErrorMsg = message
	return r
}

// SetReason sets the cause of the error for debugging.
func (r *Rerror) SetReason(reason string) *Rerror {
	r.Data.Reason = reason
	return r
}

// String prints error info.
func (r *Rerror) String() string {
	if r == nil {
		return "<nil>"
	}
	b, _ := r.MarshalJSON()
	return goutil.BytesToString(b)
}

// MarshalJSON marshals Rerror into JSON, implements json.Marshaler interface.
func (r *Rerror) MarshalJSON() ([]byte, error) {
	if r == nil {
		return []byte{}, nil
	}
	var b = append(reA, strconv.FormatInt(int64(r.ErrorCode), 10)...)
	if len(r.ErrorMsg) > 0 {
		b = append(b, reB...)
		b = append(b, utils.ToJsonStr(goutil.StringToBytes(r.ErrorMsg), false)...)
	}
	if len(r.Data.Reason) > 0 {
		b = append(b, reC...)
		b = append(b, utils.ToJsonStr(goutil.StringToBytes(r.Data.Reason), false)...)
		b = append(b, '}')
	} else {
		b = append(b, reD...)
	}
	b = append(b, '}')
	return b, nil
}

// UnmarshalJSON unmarshals a JSON description of self.
func (r *Rerror) UnmarshalJSON(b []byte) error {
	if r == nil {
		return nil
	}
	s := goutil.BytesToString(b)
	r.ErrorCode = int32(gjson.Get(s, "error_code").Int())
	r.ErrorMsg = gjson.Get(s, "error_msg").String()
	r.Data.Reason = gjson.Get(s, "data.reason").String()
	return nil
}

func hasRerror(meta *utils.Args) bool {
	return meta.Has(MetaRerror)
}

func getRerrorBytes(meta *utils.Args) []byte {
	return meta.Peek(MetaRerror)
}

// ToError converts to error
func (r *Rerror) ToError() error {
	if r == nil {
		return nil
	}
	return (*rerror)(unsafe.Pointer(r))
}

// ToRerror converts error to *Rerror
func ToRerror(err error) *Rerror {
	if err == nil {
		return nil
	}
	r, ok := err.(*rerror)
	if ok {
		return r.toRerror()
	}
	rerr := rerrUnknownError.Copy().SetReason(err.Error())
	return rerr
}

type rerror Rerror

func (r *rerror) Error() string {
	b, _ := r.toRerror().MarshalJSON()
	return goutil.BytesToString(b)
}

func (r *rerror) toRerror() *Rerror {
	return (*Rerror)(unsafe.Pointer(r))
}
