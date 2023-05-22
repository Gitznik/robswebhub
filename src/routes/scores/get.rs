use crate::routes::{
    flash_messages_utils::flash_messages_section, routing_utils::see_other_error,
    scores::post::get_match_information,
};
use actix_web::{get, web, HttpResponse};
use actix_web_flash_messages::IncomingFlashMessages;
use serde::Deserialize;
use sqlx::{query_as, types::chrono::NaiveDate, PgPool};
use uuid::Uuid;

use crate::html_base::compose_html;

#[derive(Debug, Deserialize)]
struct QueryData {
    matchup_id: Option<Uuid>,
}

#[get("/scores")]
async fn add_scores(
    query: web::Query<QueryData>,
    pg_pool: web::Data<PgPool>,
    flash_messages: IncomingFlashMessages,
) -> actix_web::Result<HttpResponse> {
    let error_html = flash_messages_section(flash_messages)
        .map_err(|e| actix_web::error::ErrorInternalServerError(e))?;
    let scores = match query.matchup_id {
        Some(matchup_id) => match_summary(matchup_id, &pg_pool)
            .await
            .map_err(|e| see_other_error("/scores", Some(e)))?,
        None => "".to_owned(),
    };
    let insert_score_form = insert_score_form(query.matchup_id);
    let main_div = include_str!("get.html");
    let main_div = format!(
        "{}\n<main class=\"container\">{}{}{}</main>",
        &error_html, &main_div, &insert_score_form, &scores
    );
    let html = compose_html(&main_div);
    Ok(HttpResponse::Ok().body(html))
}

async fn match_summary(match_id: Uuid, pg_pool: &PgPool) -> Result<String, anyhow::Error> {
    get_match_information(match_id, pg_pool).await?;
    let match_scores = get_match_scores(match_id, pg_pool).await?;
    let match_rows: Vec<String> = match_scores
        .into_iter()
        .map(|res| {
            format!(
                r#"
            <tr>
              <th scope="row">1</th>
              <td>{}</td>
              <td>{}</td>
              <td>{}</td>
              <td>{}</td>
            </tr>
        "#,
                res.winner, res.winner_score, res.loser_score, res.played_at
            )
        })
        .collect();

    Ok(format!(
        r#"
        <h1>Match Scores for Match</h1>
        <h2>{}</h2>
        <table role="grid">
          <thead>
            <tr>
              <th scope="col">#</th>
              <th scope="col">winner</th>
              <th scope="col">winner_score</th>
              <th scope="col">loser_score</th>
              <th scope="col">played_at</th>
            </tr>
          </thead>
          <tbody>
            {}
          </tbody>
        </table>
      </main>
    "#,
        match_id,
        match_rows.join("\n")
    ))
}

#[derive(Debug)]
#[allow(dead_code)]
struct MatchScore {
    match_id: Uuid,
    game_id: Uuid,
    winner: String,
    played_at: NaiveDate,
    winner_score: i16,
    loser_score: i16,
}

async fn get_match_scores(
    matchup_id: Uuid,
    pg_pool: &PgPool,
) -> Result<Vec<MatchScore>, anyhow::Error> {
    let scores = query_as!(
        MatchScore,
        r#"
        select match_id, game_id, winner, played_at, winner_score, loser_score
        from scores
        where match_id = $1
        order by played_at desc
        "#,
        matchup_id
    )
    .fetch_all(pg_pool)
    .await?;
    Ok(scores)
}

fn insert_score_form(matchup_id: Option<Uuid>) -> String {
    let default_matchup = match matchup_id {
        Some(matchup_id) => format!("value={}", matchup_id),
        None => "".to_owned(),
    };
    format!(
        r#"
        <div>
          <h2>Add Score</h2>
          <div class="grid">
            <div>
              <div class="grid">
                <h3>Single Result</h3>
                <button type="submit" form="single_result">Submit</button>
              </div>
              <form id="single_result" action="/scores" method="post">
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
            </div>
            <div>
              <div class="grid">
                <h3>Batch Upload</h3>
                <button type="submit" form="batch_upload">Submit</button>
              </div>
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
          </div>
        </div>
    "#,
        default_matchup, default_matchup
    )
}
