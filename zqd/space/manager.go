package space

import (
	"context"
	"path"
	"sync"

	"github.com/brimsec/zq/pkg/iosrc"
	"github.com/brimsec/zq/zqd/api"
	"github.com/brimsec/zq/zqd/storage"
	"github.com/brimsec/zq/zqe"
	"go.uber.org/zap"
)

type Manager struct {
	rootPath iosrc.URI
	spacesMu sync.Mutex
	spaces   map[api.SpaceID]Space
	names    map[string]api.SpaceID
	logger   *zap.Logger
}

func NewManager(root iosrc.URI, logger *zap.Logger) (*Manager, error) {
	return NewManagerWithContext(context.Background(), root, logger)
}

func NewManagerWithContext(ctx context.Context, root iosrc.URI, logger *zap.Logger) (*Manager, error) {
	mgr := &Manager{
		rootPath: root,
		spaces:   make(map[api.SpaceID]Space),
		names:    make(map[string]api.SpaceID),
		logger:   logger,
	}

	list, err := iosrc.ReadDir(ctx, root)
	if err != nil {
		return nil, err
	}
	for _, l := range list {
		if !l.IsDir() {
			continue
		}
		dir := root.AppendPath(l.Name())
		config, err := mgr.loadConfig(ctx, dir)
		if err != nil {
			if zqe.IsNotFound(err) {
				logger.Debug("Config file not found", zap.String("uri", dir.String()))
			} else {
				logger.Warn("Error loading space", zap.String("uri", dir.String()), zap.Error(err))
			}
			continue
		}

		spaces, err := loadSpaces(ctx, dir, config, mgr.logger)
		if err != nil {
			return nil, err
		}
		for _, s := range spaces {
			mgr.spaces[s.ID()] = s
			mgr.names[s.Name()] = s.ID()
		}
	}

	return mgr, nil
}

func (m *Manager) Create(ctx context.Context, req api.SpacePostRequest) (Space, error) {
	m.spacesMu.Lock()
	defer m.spacesMu.Unlock()
	if req.Name == "" && req.DataPath == "" {
		return nil, zqe.E(zqe.Invalid, "must supply non-empty name or dataPath")
	}
	var datapath iosrc.URI
	if req.DataPath != "" {
		var err error
		datapath, err = iosrc.ParseURI(req.DataPath)
		if err != nil {
			return nil, err
		}
	}
	// If name is not set then derive name from DataPath, removing and
	// replacing invalid characters.
	if req.Name == "" {
		req.Name = safeName(path.Base(datapath.Path))
		req.Name = uniqueName(m.names, req.Name)
	}
	if err := validateName(m.names, req.Name); err != nil {
		return nil, err
	}
	var storecfg storage.Config
	if req.Storage != nil {
		storecfg = *req.Storage
	}
	if storecfg.Kind == storage.UnknownStore {
		storecfg.Kind = storage.FileStore
	}
	if storecfg.Kind == storage.FileStore && m.rootPath.Scheme != "file" {
		return nil, zqe.E(zqe.Invalid, "cannot create file storage space on non-file backed data path")
	}
	id := newSpaceID()
	path := m.rootPath.AppendPath(string(id))
	src, err := iosrc.GetSource(path)
	if err != nil {
		return nil, err
	}
	if dirmk, ok := src.(iosrc.DirMaker); ok {
		if err := dirmk.MkdirAll(path, 0754); err != nil {
			return nil, err
		}
	}
	if req.DataPath == "" {
		datapath = path
	}
	conf := config{Version: configVersion, Name: req.Name, DataURI: datapath, Storage: storecfg}
	if err := writeConfig(path, conf); err != nil {
		iosrc.RemoveAll(context.Background(), path)
		return nil, err
	}
	spaces, err := loadSpaces(ctx, path, conf, m.logger)
	if err != nil {
		return nil, err
	}
	s := spaces[0]
	m.spaces[s.ID()] = s
	m.names[s.Name()] = s.ID()
	return s, err
}

func (m *Manager) CreateSubspace(ctx context.Context, parent Space, req api.SubspacePostRequest) (Space, error) {
	m.spacesMu.Lock()
	defer m.spacesMu.Unlock()
	if err := validateName(m.names, req.Name); err != nil {
		return nil, err
	}
	as, ok := parent.(*archiveSpace)
	if !ok {
		return nil, zqe.E(zqe.Invalid, "space does not support creating subspaces")
	}

	s, err := as.CreateSubspace(ctx, req)
	if err != nil {
		return nil, err
	}
	m.spaces[s.ID()] = s
	m.names[s.Name()] = s.ID()
	return s, nil
}

func (m *Manager) UpdateSpace(space Space, req api.SpacePutRequest) error {
	m.spacesMu.Lock()
	defer m.spacesMu.Unlock()
	if err := validateName(m.names, req.Name); err != nil {
		return err
	}

	// Right now you can only update a name in a SpacePutRequest but eventually
	// there will be other options.
	oldname := space.Name()
	if oldname == req.Name {
		return nil
	}
	if err := space.update(req); err != nil {
		return err
	}
	delete(m.names, oldname)
	m.names[space.Name()] = space.ID()
	return nil
}

func (m *Manager) Get(id api.SpaceID) (Space, error) {
	m.spacesMu.Lock()
	defer m.spacesMu.Unlock()

	space, exists := m.spaces[id]
	if !exists {
		return nil, ErrSpaceNotExist
	}

	return space, nil
}

func (m *Manager) Delete(ctx context.Context, id api.SpaceID) error {
	space, err := m.Get(id)
	if err != nil {
		return err
	}

	m.spacesMu.Lock()
	defer m.spacesMu.Unlock()
	name := space.Name()
	if err := space.delete(ctx); err != nil {
		return err
	}

	delete(m.spaces, id)
	delete(m.names, name)
	return nil
}

func (m *Manager) List(ctx context.Context) ([]api.SpaceInfo, error) {
	result := []api.SpaceInfo{}

	m.spacesMu.Lock()
	defer m.spacesMu.Unlock()
	for id := range m.spaces {
		sp := m.spaces[id]
		info, err := sp.Info(ctx)
		if err != nil {
			// XXX should add ability to derive request id from context if it
			// exists for current ctx.
			m.logger.Warn("Could not get space info",
				zap.String("space_id", string(id)),
				zap.Error(err),
			)
			return nil, err
		}
		result = append(result, info)
	}
	return result, nil
}
