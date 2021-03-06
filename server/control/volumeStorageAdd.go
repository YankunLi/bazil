package control

import (
	"log"

	"bazil.org/bazil/db"
	"bazil.org/bazil/server/control/wire"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func (c controlRPC) VolumeStorageAdd(ctx context.Context, req *wire.VolumeStorageAddRequest) (*wire.VolumeStorageAddResponse, error) {
	if err := c.app.ValidateKV(req.Backend); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	addStorage := func(tx *db.Tx) error {
		vol, err := tx.Volumes().GetByName(req.VolumeName)
		if err != nil {
			return err
		}
		sharingKey, err := tx.SharingKeys().Get(req.SharingKeyName)
		if err != nil {
			return err
		}
		return vol.Storage().Add(req.Name, req.Backend, sharingKey)
	}
	if err := c.app.DB.Update(addStorage); err != nil {
		switch err {
		case db.ErrVolNameNotFound:
			return nil, grpc.Errorf(codes.FailedPrecondition, "%v", err)
		case db.ErrSharingKeyNameInvalid:
			return nil, grpc.Errorf(codes.InvalidArgument, "%v", err)
		case db.ErrSharingKeyNotFound:
			return nil, grpc.Errorf(codes.FailedPrecondition, "%v", err)
		case db.ErrVolumeStorageExist:
			return nil, grpc.Errorf(codes.AlreadyExists, err.Error())
		}
		log.Printf("db update error: add storage %q: %v", req.Name, err)
		return nil, grpc.Errorf(codes.Internal, "Internal error")
	}
	return &wire.VolumeStorageAddResponse{}, nil
}
