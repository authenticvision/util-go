package av

import (
	"encoding/json"
	"strconv"
	"strings"
)

type SLID int64

const InvalidSLID SLID = 0

func (slid SLID) String() string {
	return slid.Base36()
}

func (slid SLID) Base10() string {
	return strconv.FormatUint(uint64(slid), 10)
}

func (slid SLID) Base36() string {
	return strings.ToUpper(strconv.FormatUint(uint64(slid), 36))
}

// SLID36 is encoded/decoded as base36 in JSON
type SLID36 SLID

func (slid SLID36) SLID() SLID {
	return SLID(slid)
}

func (slid *SLID36) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	i, err := strconv.ParseUint(s, 36, 64)
	if err != nil {
		return err
	}
	*slid = SLID36(i)
	return nil
}

func (slid SLID36) MarshalJSON() ([]byte, error) {
	return json.Marshal(slid.SLID().Base36())
}
