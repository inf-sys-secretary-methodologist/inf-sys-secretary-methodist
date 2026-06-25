package entities

// ODataStudentDebt represents an academic-debt record from the 1С OData
// response (the Catalog_АкадемическиеЗадолженности catalog). It is a pure
// wire DTO: unlike employees and students, debts are not persisted inside the
// integration module — they flow straight into the student_debts bounded
// context, whose 1С adapter maps this struct onto its own ImportedDebt port
// (including the Russian-label → wire-code control-form translation). Keeping
// the mapping out of this package preserves the bounded-context boundary.
//
// Field tags mirror the Cyrillic 1С property names, matching the convention
// established by ODataStudent / ODataEmployee.
type ODataStudentDebt struct {
	RefKey       string `json:"Ref_Key"`                 // 1С GUID — stable source reference
	DataVersion  string `json:"DataVersion"`             // 1С row version
	DeletionMark bool   `json:"DeletionMark"`            // soft-delete flag in 1С
	StudentRef   string `json:"Студент_Key,omitempty"`   // GUID link to the student catalog
	StudentName  string `json:"Студент,omitempty"`       // denormalized full name
	GroupName    string `json:"Группа,omitempty"`        // academic group
	Discipline   string `json:"Дисциплина,omitempty"`    // discipline the debt is owed against
	Semester     int    `json:"Семестр,omitempty"`       // 1..12
	ControlForm  string `json:"ФормаКонтроля,omitempty"` // Russian label, e.g. "Экзамен"/"Зачёт"
}
