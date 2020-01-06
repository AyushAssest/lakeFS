package operations

import (
	"fmt"
	"treeverse-lake/db"
	"treeverse-lake/gateway/errors"
	"treeverse-lake/gateway/permissions"
	"treeverse-lake/gateway/serde"

	"golang.org/x/xerrors"
)

type HeadObject struct{}

func (controller *HeadObject) GetArn() string {
	return "arn:treeverse:repos:::{bucket}"
}

func (controller *HeadObject) GetPermission() string {
	return permissions.PermissionReadRepo
}

func (controller *HeadObject) Handle(o *PathOperation) {
	obj, err := o.Index.ReadObject(o.Repo, o.Branch, o.Path)
	if xerrors.Is(err, db.ErrNotFound) {
		// TODO: create distinction between missing repo & missing key
		o.Log().
			WithField("path", o.Path).
			WithField("branch", o.Branch).
			WithField("repo", o.Repo).
			WithError(err).
			Error("path not found")
		o.EncodeError(errors.Codes.ToAPIErr(errors.ErrNoSuchKey))
		return
	}
	if err != nil {
		o.Log().WithError(err).Error("failed querying path")
		o.EncodeError(errors.Codes.ToAPIErr(errors.ErrInternalError))
		return
	}
	o.SetHeader("Accept-Ranges", "bytes")
	o.SetHeader("Last-Modified", serde.HeaderTimestamp(obj.GetTimestamp()))
	o.SetHeader("ETag", fmt.Sprintf("\"%s\"", obj.GetBlob().GetChecksum()))
	o.SetHeader("Content-Length", fmt.Sprintf("%d", obj.GetSize()))
}