package network

type GetStatusMessage struct {

}

type StatusMessage struct {
	Version uint32
	CurrentHeight uint32
}
