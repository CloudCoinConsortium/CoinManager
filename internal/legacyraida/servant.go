package legacyraida

type Servant struct {
	Raida       *LegacyRAIDA
  progressChannel chan interface{}
}

type Error struct {
	Message string
}

func NewServant(progressChannel chan interface{}) *Servant {
	//fmt.Println("new servant")
	Raida := New(progressChannel)

	return &Servant{
		Raida:       Raida,
    progressChannel: progressChannel,
	}
}


