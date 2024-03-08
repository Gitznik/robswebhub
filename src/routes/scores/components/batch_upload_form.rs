use actix_web::{get, web, HttpResponse};
use serde::Deserialize;
use uuid::Uuid;

#[derive(Debug, Deserialize)]
struct QueryData {
    matchup_id: Option<Uuid>,
}

#[get("/batch_result_form")]
async fn get_batch_result_form(query: web::Query<QueryData>) -> actix_web::Result<HttpResponse> {
    let default_matchup = match query.matchup_id {
        Some(matchup_id) => format!("value={}", matchup_id),
        None => "".to_owned(),
    };
    let body = format!(
        r#"
    <div id="score_entry_form">
        <form id="batch_upload" action="/scores_batch" method="post">
            <div class="grid">
              <label for="matchup_id">
                Matchup Id
                <input type="text" id="matchup_id" name="matchup_id" placeholder="Matchup Id" {} required>
              </label>
              <label for="raw_matches_list">
                Raw matches list
                <textarea id="raw_matches_list" name="raw_matches_list" placeholder="Raw matches list, e.g. 2023-02-22 P1 2:1" rows="5" required></textarea>
              </label>
            </div>
        </form>
    </div>
    "#,
        default_matchup
    );
    Ok(HttpResponse::Ok().body(body))
}
