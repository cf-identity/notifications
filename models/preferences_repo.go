package models

type PreferencesRepo struct{}

type PreferencesRepoInterface interface {
    FindNonCriticalPreferences(ConnectionInterface, string) ([]Preference, error)
}

func NewPreferencesRepo() PreferencesRepo {
    return PreferencesRepo{}
}

func (repo PreferencesRepo) FindNonCriticalPreferences(conn ConnectionInterface, userGUID string) ([]Preference, error) {
    preferences := []Preference{}

    sql := `SELECT kinds.id as kind_id, kinds.client_id as client_id
            FROM kinds LEFT OUTER JOIN receipts ON kinds.id = receipts.kind_id
            WHERE kinds.critical = "false"
            AND kinds.client_id IN
                (SELECT DISTINCT kinds.client_id
                 FROM kinds JOIN receipts ON kinds.client_id = receipts.client_id
                 WHERE receipts.user_guid = ?)`

    _, err := conn.Select(&preferences, sql, userGUID)
    if err != nil {
        return preferences, err
    }

    for index, _ := range preferences {
        preferences[index].Email = "true"
    }

    return preferences, nil
}
