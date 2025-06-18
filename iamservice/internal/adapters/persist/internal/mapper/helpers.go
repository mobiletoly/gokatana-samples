package mapper

import (
	"fmt"
	"github.com/mobiletoly/gokatana/katapp"
	"strconv"
)

func ModelIdToRepoId(ID string) (int64, error) {
	repoID, err := strconv.ParseInt(ID, 10, 32)
	if err != nil {
		return 0, katapp.NewErr(katapp.ErrInvalidInput, fmt.Sprintf("failed to parse record id: %s", ID))
	}
	return repoID, nil
}

func RepoIdToModelId(ID int64) *string {
	strID := strconv.FormatInt(ID, 10)
	return &strID
}
