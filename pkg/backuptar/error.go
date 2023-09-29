package backuptar

import "errors"

var ErrPrepareToAppend = errors.New("tar file is not prepared to append")
