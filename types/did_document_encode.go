package types

func (d *DIDDocument) unpack(
	context []string, id string,
) error {
	d.context_ = context

	did, err := NewDIDRefFromString(id)
	if err != nil {
		return err
	}
	d.id = *did

	return nil
}

func (d *Service) unpack(id, svcType, svcEP string) error {
	did, err := NewDIDURLRefFromString(id)
	if err != nil {
		return err
	}
	d.id = *did
	d.serviceType = svcType
	d.serviceEndPoint = svcEP

	return nil
}
