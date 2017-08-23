package service

import (
	"time"

	"errors"
	"sync"

	"github.com/dedis/cothority/skipchain"
	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/protocol"

	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/dedis/onet.v1/network"
)

var templateID onet.ServiceID

func init() {
	templateID, _ = onet.RegisterNewService(api.ServiceName, newService)
	network.RegisterMessage(&storage{})
	network.RegisterMessage(&Base{})
	network.RegisterMessage(&Ballot{})
	network.RegisterMessage(&Config{})
}

// Service is our template-service
type Service struct {
	*onet.ServiceProcessor

	storage *storage
}

// storageID reflects the data we're storing - we could store more
// than one structure.
const storageID = "main"

// storage is used to save our data.
type storage struct {
	sync.Mutex

	Elections map[string]*Election
}

type Base struct {
	Key abstract.Point
}

type Election struct {
	Genesis *skipchain.SkipBlock
	Latest  *skipchain.SkipBlock

	*protocol.SharedSecret
}

type Ballot struct {
	Data string
}

type Config struct {
	Name    string
	Genesis *skipchain.SkipBlock
}

// GenerateRequest ...
func (service *Service) GenerateRequest(request *api.GenerateRequest) (
	*api.GenerateResponse, onet.ClientError) {

	length := len(request.Roster.List)
	tree := request.Roster.GenerateNaryTreeWithRoot(length, service.ServerIdentity())
	dkg, err := service.CreateProtocol(protocol.NameDKG, tree)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	client := skipchain.NewClient()
	genesis, err := client.CreateGenesis(request.Roster, 1, 1,
		skipchain.VerificationNone, nil, nil)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	config, _ := network.Marshal(&Config{Name: request.Name, Genesis: genesis})
	setupDKG := dkg.(*protocol.SetupDKG)
	setupDKG.Wait = true
	if err = setupDKG.SetConfig(&onet.GenericConfig{Data: config}); err != nil {
		return nil, onet.NewClientError(err)
	}

	if err := dkg.Start(); err != nil {
		return nil, onet.NewClientError(err)
	}

	select {
	case <-setupDKG.Done:
		shared, _ := setupDKG.SharedSecret()
		service.storage.Lock()
		service.storage.Elections[request.Name] = &Election{genesis, genesis, shared}
		service.storage.Unlock()
		service.save()

		return &api.GenerateResponse{Key: shared.X, Hash: genesis.Hash}, nil
	case <-time.After(2000 * time.Millisecond):
		return nil, onet.NewClientError(errors.New("dkg didn't finish in time"))
	}
}

func (service *Service) NewProtocol(node *onet.TreeNodeInstance, conf *onet.GenericConfig) (
	onet.ProtocolInstance, error) {
	switch node.ProtocolName() {
	case protocol.NameDKG:
		dkg, err := protocol.NewSetupDKG(node)
		if err != nil {
			return nil, err
		}

		setupDKG := dkg.(*protocol.SetupDKG)
		go func(conf *onet.GenericConfig) {
			<-setupDKG.Done
			shared, err := setupDKG.SharedSecret()
			if err != nil {
				return
			}

			_, data, err := network.Unmarshal(conf.Data)
			if err != nil {
				return
			}

			config := data.(*Config)

			service.storage.Lock()
			election := &Election{config.Genesis, config.Genesis, shared}
			service.storage.Elections[config.Name] = election
			service.storage.Unlock()
			service.save()
		}(conf)

		return dkg, nil
	default:
		return nil, errors.New("Unknown protocol")
	}
}

func (service *Service) CastRequest(request *api.CastRequest) (
	*api.CastResponse, onet.ClientError) {

	election, found := service.storage.Elections[request.Name]
	if !found {
		return nil, onet.NewClientError(errors.New("Election not found"))
	}

	client := skipchain.NewClient()
	response, err := client.StoreSkipBlock(election.Latest, nil, []byte(request.Ballot))
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	service.storage.Lock()
	election.Latest = response.Latest
	service.storage.Unlock()
	service.save()

	return &api.CastResponse{}, nil
}

// saves all skipblocks.
func (s *Service) save() {
	s.storage.Lock()
	defer s.storage.Unlock()
	err := s.Save(storageID, s.storage)
	if err != nil {
		log.Error("Couldn't save file:", err)
	}
}

// Tries to load the configuration and updates the data in the service
// if it finds a valid config-file.
func (s *Service) tryLoad() error {
	s.storage = &storage{Elections: make(map[string]*Election)}
	if !s.DataAvailable(storageID) {
		return nil
	}

	msg, err := s.Load(storageID)
	if err != nil {
		return err
	}
	var ok bool
	s.storage, ok = msg.(*storage)
	if !ok {
		return errors.New("Data of wrong type")
	}
	return nil
}

func newService(c *onet.Context) onet.Service {
	s := &Service{ServiceProcessor: onet.NewServiceProcessor(c)}

	if err := s.RegisterHandlers(s.GenerateRequest, s.CastRequest); err != nil {
		log.ErrFatal(err, "Couldn't register messages")
	}

	if err := s.tryLoad(); err != nil {
		log.Error(err)
	}

	return s
}
