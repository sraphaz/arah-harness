package core

// TaskService is the application use-case layer over StateStore + choreography.
type TaskService struct {
	Store  StateStore
	Router ChoreographyResolver
}

func (s *TaskService) Create(objective, area string, wc WorkClass, intent IntentType) (*Contract, string, error) {
	if area == "" {
		area = "backend"
	}
	if wc == "" {
		wc = WorkStandard
	}
	if intent == "" {
		intent = IntentExecution
	}
	routing, err := s.Router.Resolve(area, "")
	if err != nil {
		return nil, "", err
	}
	c, err := NewContract(objective, area, wc, intent, routing)
	if err != nil {
		return nil, "", err
	}
	if err := c.Start(); err != nil {
		return nil, "", err
	}
	path, err := s.Store.Save(c)
	if err != nil {
		return nil, "", wrapStore(err)
	}
	return c, path, nil
}

func (s *TaskService) Get(taskID string) (*Contract, string, error) {
	c, path, err := s.Store.Get(taskID)
	if err != nil {
		return nil, "", err
	}
	return c, path, nil
}

func (s *TaskService) Complete(taskID string, evidence []string) (*Contract, string, error) {
	c, _, err := s.Store.Get(taskID)
	if err != nil {
		return nil, "", err
	}
	if err := c.Complete(evidence); err != nil {
		return nil, "", err
	}
	path, err := s.Store.Save(c)
	if err != nil {
		return nil, "", wrapStore(err)
	}
	return c, path, nil
}

func (s *TaskService) Block(taskID, reason string) (*Contract, string, error) {
	c, _, err := s.Store.Get(taskID)
	if err != nil {
		return nil, "", err
	}
	if err := c.Block(reason); err != nil {
		return nil, "", err
	}
	path, err := s.Store.Save(c)
	if err != nil {
		return nil, "", wrapStore(err)
	}
	return c, path, nil
}

func wrapStore(err error) error {
	return errf("STATE.STORE_ERROR", err.Error(), nil)
}
