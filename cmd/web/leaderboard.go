package main

import (
	"fmt"
	"html/template"
	"io"

	"github.com/gocasters/rankr/adapter/leaderboardstat"
)

type leaderboardView struct {
	ProjectID string
	Rows      []leaderboardstat.PublicLeaderboardRow
	Error     bool
}

const leaderboardTemplate = `<!doctype html>
<html lang="fa" dir="rtl">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>لیدربورد پروژه {{.ProjectID}} | Rankr</title>
	<style>
		body {
			font-family: Tahoma, sans-serif;
			max-width: 60rem;
			margin: 2rem auto;
			padding: 0 1rem;
			background: #f5f8fb;
			color: #183047;
		}
		main {
			background: white;
			padding: 1.5rem;
			border-radius: 12px;
		}
		.table-wrapper { overflow-x: auto; }
		table { width: 100%; border-collapse: collapse; }
		th, td {
			text-align: start;
			padding: .75rem;
			border-bottom: 1px solid #d5e0e8;
		}
		caption { text-align: start; margin-bottom: 1rem; }
	</style>
</head>
<body>
	<main>
		<h1>لیدربورد عمومی</h1>
		<p>شناسهٔ مخزن: {{.ProjectID}}</p>

		{{if .Error}}
			<p role="alert">دریافت اطلاعات انجام نشد.</p>
		{{else if .Rows}}
			<div class="table-wrapper">
				<table>
					<caption>امتیاز مشارکت‌کنندگان</caption>
					<thead>
						<tr>
							<th scope="col">رتبه</th>
							<th scope="col">شناسهٔ کاربر</th>
							<th scope="col">امتیاز</th>
						</tr>
					</thead>
					<tbody>
						{{range .Rows}}
						<tr>
							<td>{{.Rank}}</td>
							<td>{{.UserID}}</td>
							<td>{{.Score}}</td>
						</tr>
						{{end}}
					</tbody>
				</table>
			</div>
		{{else}}
			<p>هنوز امتیازی ثبت نشده است.</p>
		{{end}}
	</main>
</body>
</html>`

func renderLeaderboard(w io.Writer, view leaderboardView) error {
	page, err := template.New("leaderboard").Parse(leaderboardTemplate)
	if err != nil {
		return fmt.Errorf("web: parse leaderboard template: %w", err)
	}

	if err := page.Execute(w, view); err != nil {
		return fmt.Errorf("web: render leaderboard: %w", err)
	}

	return nil
}
