package geecaches

import pb "geecache-s/geecachespb"

type PeerPicker interface {
	PickPeer(key string) (PeerGetter, bool)
	SelfAddr() string
}

type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error
}
