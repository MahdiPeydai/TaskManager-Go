package models

type User struct {
	BaseModel
	Username       string  `gorm:"column:username;size:20;unique; not null;unique"`
	Firstname      *string `gorm:"column:first_name;size:15;null"`
	Lastname       *string `gorm:"column:last_name;size:25;null"`
	Email          *string `gorm:"column:email;size:64;unique;null"`
	MobileNumber   *string `gorm:"column:mobile_number;size:11;unique;null"`
	Password       *string `gorm:"column:password;size:64;null;"`
	IsActive       bool    `gorm:"column:is_active;default:true;"`
	MobileVerified bool    `gorm:"column:mobile_verified;default:false;"`
	Roles          []UserRole
}

type Role struct {
	BaseModel
	Name  string `gorm:"size:15;unique;not null;unique"`
	Users []UserRole
}

type UserRole struct {
	BaseModel
	User   *User `gorm:"foreignKey:UserId;constraint:OnUpdate:No Action,OnDelete:No Action;"`
	UserId int
	Role   *Role `gorm:"foreignKey:RoleId;constraint:OnUpdate:No Action,OnDelete:No Action;"`
	RoleId int
}
