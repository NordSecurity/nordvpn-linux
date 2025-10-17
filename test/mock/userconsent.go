package mock

type AnalyticsConsentCheckerMock struct {
	ConsentCompleted bool
}

func (*AnalyticsConsentCheckerMock) PrepareDaemonIfConsentNotCompleted() {}

func (c *AnalyticsConsentCheckerMock) IsConsentFlowCompleted() bool {
	return c.ConsentCompleted
}
