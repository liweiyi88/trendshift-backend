package dbutils

import "testing"

func TestQueryBuilder(t *testing.T) {
	qb := NewQueryBuilder()

	qb.Query("select repositories.id, count(*) as count from repositories join trending_repositories on repositories.id = trending_repositories.repository_id")
	qb.Where("`trending_repositories`.`language` = ?", "PHP")
	qb.Where("`trending_repositories`.`trend_date` = ?", "2023-09-09")
	qb.GroupBy("repositories.id")
	qb.OrderBy("count", "DESC")
	qb.OrderBy("repositories.id", "ASC")

	query, args := qb.GetQuery()

	want := "select repositories.id, count(*) as count from repositories join trending_repositories on repositories.id = trending_repositories.repository_id WHERE `trending_repositories`.`language` = ? AND `trending_repositories`.`trend_date` = ? GROUP BY repositories.id ORDER BY count DESC, repositories.id ASC"
	if query != want {
		t.Errorf("want: %s, but got: %s", want, query)
	}

	if args[0] != "PHP" || args[1] != "2023-09-09" {
		t.Errorf("unexpected args: %v", args...)
	}

	query, args = qb.Limit(10).GetQuery()
	want = "select repositories.id, count(*) as count from repositories join trending_repositories on repositories.id = trending_repositories.repository_id WHERE `trending_repositories`.`language` = ? AND `trending_repositories`.`trend_date` = ? GROUP BY repositories.id ORDER BY count DESC, repositories.id ASC LIMIT ?"
	if query != want {
		t.Errorf("want: %s, but got: %s", want, query)
	}

	if args[0] != "PHP" || args[1] != "2023-09-09" || args[2] != 10 {
		t.Errorf("unexpected args: %v", args...)
	}
}

func TestReset(t *testing.T) {
	qb := NewQueryBuilder()

	qb.Query("select repositories.id, count(*) as count from repositories join trending_repositories on repositories.id = trending_repositories.repository_id")
	qb.Where("`trending_repositories`.`language` = ?", "PHP")
	qb.Where("`trending_repositories`.`trend_date` = ?", "2023-09-09")
	qb.GroupBy("repositories.id")
	qb.OrderBy("count", "DESC")

	qb.reset()

	if qb.args != nil || qb.criteria != nil || qb.groupBy != "" || qb.limit != "" || qb.orderBy != nil || qb.query != "" {
		t.Errorf("query builder has not been rest: %+v", qb)
	}
}
