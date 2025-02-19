package geecaches

import pb "geecache-s/lib/geecachespb"

type PeerPicker interface {
	PickPeer(key string) (PeerHandler, bool)
	SelfAddr() string
}

type PeerHandler interface {
	Get(in *pb.GetRequest, out *pb.GetResponse) error
	Add(in *pb.AddRequest, out *pb.Empty) error
}
