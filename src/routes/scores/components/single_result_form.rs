use actix_web::{get, web, HttpResponse};
use serde::Deserialize;
use uuid::Uuid;

#[derive(Debug, Deserialize)]
struct QueryData {
    matchup_id: Option<Uuid>,
}

#[get("/single_result_form")]
async fn get_single_result_form(query: web::Query<QueryData>) -> actix_web::Result<HttpResponse> {
    let default_matchup = match query.matchup_id {
        Some(matchup_id) => format!("value={}", matchup_id),
        None => "".to_owned(),
    };
    let body = format!(
        r#"
            <div id="score_entry_form">
              <form id="single_result" action="/scores/scores_single" method="post">
                <div class="grid">
                  <label for="matchup_id">
                    Matchup Id
                    <input type="text" id="matchup_id" name="matchup_id" placeholder="Matchup Id" {} required>
                  </label>
                  <label for="winner_initials">
                    Winner Credentials
                    <input type="text" id="winner_initials" name="winner_initials" placeholder="Winner Initials" required>
                  </label>
                </div>
                <div class="grid">
                  <label for="score">
                    Score, separated by ":"
                    <input type="text" id="score" name="score" placeholder="Score" required>
                  </label>
                  <label for="played_at">
                    Date the match was played at
                    <input type="date" id="played_at" name="played_at" placeholder="dd.mm.yyyy" required>
                  </label>
                </div>
              </form>
              <div class="grid">
                <button type="submit" form="single_result">Submit</button>
              </div>
        "#,
        default_matchup
    );
    Ok(HttpResponse::Ok().body(body))
}
