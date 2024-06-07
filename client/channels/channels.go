package channels

// CHMessageFields struct to send message.
type CHMessageFields[T any] struct {
	Resp T
	ID   string
}

// ChannelStruct struct to use channels.
type ChannelStruct[T any] struct {
	MapChannels     map[string]chan CHMessageFields[T]
	AddChannel      chan MapEntry[T]
	RemoveChannel   chan string
	ResponseChannel chan CHMessageFields[T]
}

// MapEntry struct to create channels.
type MapEntry[T any] struct {
	Id string
	Ch chan CHMessageFields[T]
}

// ProvideChannels func to initialize logic channels.
func ProvideChannels[T any]() *ChannelStruct[T] {
	channels := make(map[string]chan CHMessageFields[T])
	addChannel := make(chan MapEntry[T])
	removeChannel := make(chan string)
	responseChannel := make(chan CHMessageFields[T])

	go func() {
		for {
			select {
			case ent := <-addChannel:
				channels[ent.Id] = ent.Ch
			case resp := <-responseChannel:
				channel, ok := channels[resp.ID]
				if ok {
					channel <- resp
				}
			case id := <-removeChannel:
				delete(channels, id)
			}
		}
	}()

	mc := &ChannelStruct[T]{
		MapChannels:     channels,
		AddChannel:      addChannel,
		RemoveChannel:   removeChannel,
		ResponseChannel: responseChannel,
	}
	return mc
}

// Set to update data messages from channels.
func (m *ChannelStruct[T]) Set(msg CHMessageFields[T]) {
	go func() {
		m.ResponseChannel <- msg
	}()
}

// Delete to remove specific channel by ID.
func (m *ChannelStruct[T]) Delete(id string) {
	go func() {
		m.RemoveChannel <- id
	}()
}

// Init to create channel.
func (m *ChannelStruct[T]) Init(id string) chan CHMessageFields[T] {
	ch := make(chan CHMessageFields[T])
	entry := MapEntry[T]{Id: id, Ch: ch}
	m.AddChannel <- entry
	return ch
}

// CloseChannels to close all channels.
func (m *ChannelStruct[T]) CloseChannels() {
	go func() {
		close(m.AddChannel)
		close(m.ResponseChannel)
		close(m.RemoveChannel)
	}()
}
