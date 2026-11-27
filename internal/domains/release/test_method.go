package release

// TestMethod is a QA procedure that belongs to a release.
// Production flashing stays on Filebrowser; QA reads these steps from ReleaseDesk.
type TestMethod struct {
	ID             int    `xorm:"pk autoincr" json:"id"`
	ReleaseID      int    `xorm:"not null index" json:"releaseId"`
	Title          string `xorm:"not null" json:"title"`
	Steps          string `xorm:"longtext" json:"steps"`
	ExpectedResult string `xorm:"text" json:"expectedResult"`
	SortOrder      int    `json:"sortOrder"`
}

type TestPlan struct {
	ReleaseID int          `json:"releaseId"`
	Overview  string       `json:"overview"`
	Methods   []TestMethod `json:"methods"`
}
