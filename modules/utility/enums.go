package utility

//Enums is a enum collection
var enumsInstance enum

//Enums is the global enums
func Enums() enum {
	if enumsInstance.Others.Tester != "tester" {
		enumsInstance.LoadEnums()
	}
	return enumsInstance
}

//Enum are the collection of enums :-)
type enum struct {
	//HeaderID is where the user id stored in the context
	Others other
	//ErrorMessage provides a list of error messages
	ErrorMessages errorMessage
	//Roles provides a list of roles
	Roles role
	//Status provides a list of queue status
	Status status
}

//LoadEnums initiates all global variables
func (enums *enum) LoadEnums() {
	enums.loadErrorMessageEnums()
	enums.loadRoleEnums()
	enums.loadOtherEnums()
	enums.loadStatus()
}

func (enums *enum) loadStatus() {
	enums.Status.Active = "Active"
	enums.Status.Finished = "Finished"
	enums.Status.Running = "Running"
	enums.Status.Cancelled = "Cancelled"
}

//LoadOtherEnums assign values to enums
func (enums *enum) loadOtherEnums() {
	enums.Others.Tester = "test"

}

//LoadErrorMessageEnums assign values to enums.ErrorMessages
func (enums *enum) loadErrorMessageEnums() {
	enums.ErrorMessages.AuthFailed = "Authentication failed, please check your credentials."
	enums.ErrorMessages.PageNotFound = "Page Not found."
	enums.ErrorMessages.SystemError = "System Error, please try later or contact the Administrator."
	enums.ErrorMessages.LackOfRegInfo = "Registration failed, please ensure you have provided at least Email, Password and Name."
	enums.ErrorMessages.UserExist = "User already exist."
	enums.ErrorMessages.LackOfInfo = "Fail to add an item, please ensure you have provided necessary info"
	enums.ErrorMessages.RecordExist = "Fail to add an item, the data is already exist"
	enums.ErrorMessages.RecordNotFound = "Fail to find the record"
}

//LoadRoleEnums loads a list of predefined roles
func (enums *enum) loadRoleEnums() {
	enums.Roles.Admin = "Admin"
	enums.Roles.Member = "Member"
	enums.Roles.Test = "Test"
}

//HTTPStatusStruct is the struct for http status
type hTTPStatusStruct struct {
	OK, MovedPermanently, BadRequest, Unauthorized, NotFound, Conflict, NoContent int
}

//ErrorMessage is the collection of error messages
type errorMessage struct {
	AuthFailed, PageNotFound, SystemError, LackOfRegInfo, UserExist, LackOfInfo, RecordExist, RecordNotFound string
}

//Role is the collection of roles
type role struct {
	Admin, Member, Test string
}

//status is the collection of roles
type status struct {
	Active, Running, Finished, Cancelled string
}

//Other is the struct of uncategorise enums
type other struct {
	Tester string
}
