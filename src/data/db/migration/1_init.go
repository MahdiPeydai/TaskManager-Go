package migrations

import (
	"github.com/mahdipeydai/taskmanager-go/config"
	"github.com/mahdipeydai/taskmanager-go/constants"
	"github.com/mahdipeydai/taskmanager-go/data/db"
	"github.com/mahdipeydai/taskmanager-go/data/models"
	"github.com/mahdipeydai/taskmanager-go/pkg/logging"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var logger = logging.GetLogger(config.GetConfig())

func UpInit() {
	cfg := config.GetConfig()
	logger.Info(logging.Postgres, logging.Migration, "Migration started ...", nil)

	database := db.GetDB()

	createTables(database)
	createDefaultInformation(database, cfg)

	logger.Info(logging.Postgres, logging.Migration, "Migration done.", nil)
}

func createTables(db *gorm.DB) {

	tables := make([]interface{}, 0)

	// User
	tables = addNewTable(db, models.User{}, tables)
	tables = addNewTable(db, models.Role{}, tables)
	tables = addNewTable(db, models.UserRole{}, tables)

	err := db.AutoMigrate(tables...)
	if err != nil {
		extras := map[logging.ExtraKey]interface{}{logging.ErrorMessage: err.Error()}
		logger.Fatal(logging.Postgres, logging.Migration, "Migration failed.", extras)
	}
	logger.Info(logging.Postgres, logging.Migration, "tables created", nil)
}

func addNewTable(db *gorm.DB, model interface{}, tables []interface{}) []interface{} {
	if !db.Migrator().HasTable(model) {
		tables = append(tables, model)
	}
	return tables
}

func createDefaultInformation(db *gorm.DB, cfg *config.Config) {
	adminRole := models.Role{Name: constants.AdminRoleName}
	createRoleIfExists(db, &adminRole)

	adminUser := models.User{Username: constants.AdminUsername}
	hashedAdminPass, _ := bcrypt.GenerateFromPassword([]byte(cfg.Admin.Password), bcrypt.DefaultCost)
	hashedAdminPassString := string(hashedAdminPass)
	adminUser.Password = &hashedAdminPassString
	createUserIfExists(db, &adminUser, &adminRole)

	defaultRole := models.Role{Name: constants.DefaultRoleName}
	createRoleIfExists(db, &defaultRole)
}

func createRoleIfExists(db *gorm.DB, r *models.Role) {
	exists := 0
	db.Model(r).Select("1").Where("name = ?", r.Name).First(&exists)
	if exists == 0 {
		db.Create(&r)
	}
}

func createUserIfExists(db *gorm.DB, u *models.User, ar *models.Role) {
	exists := 0
	db.Model(u).Select("1").Where("username = ?", u.Username).First(&exists)
	if exists == 0 {
		db.Create(&u)
		ur := models.UserRole{UserId: u.Id, RoleId: ar.Id}
		db.Create(&ur)
	}
}

func DownInit() error {
	logger.Info(logging.Postgres, logging.Migration, "Rollback started ...", nil)

	database := db.GetDB()

	err := database.Transaction(func(tx *gorm.DB) error {
		deleteDefaultInformation(tx)
		return dropTables(tx)
	})

	if err != nil {
		extras := map[logging.ExtraKey]interface{}{
			logging.ErrorMessage: err.Error(),
		}
		logger.Fatal(
			logging.Postgres,
			logging.Migration,
			"Rollback failed.",
			extras,
		)
	}

	logger.Info(logging.Postgres, logging.Migration, "Rollback done.", nil)
	return nil
}

func dropTables(db *gorm.DB) (err error) {
	tables := []interface{}{
		models.UserRole{},
		models.User{},
		models.Role{},
	}

	for _, table := range tables {
		if db.Migrator().HasTable(table) {
			if err := db.Migrator().DropTable(table); err != nil {
				extras := map[logging.ExtraKey]interface{}{
					logging.ErrorMessage: err.Error(),
				}
				logger.Error(
					logging.Postgres,
					logging.Migration,
					"Rollback failed.",
					extras,
				)
				return err
			}
		}
	}

	logger.Info(logging.Postgres, logging.Migration, "tables dropped", nil)
	return nil
}

func deleteDefaultInformation(db *gorm.DB) {
	db.Where("username = ?", constants.AdminUsername).
		Delete(&models.User{})

	db.Where(
		"name IN ?",
		[]string{
			constants.AdminRoleName,
			constants.DefaultRoleName,
		},
	).Delete(&models.Role{})
}
