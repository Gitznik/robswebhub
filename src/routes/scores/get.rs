use crate::routes::{routing_utils::see_other, scores::post::get_match_information};
use actix_web::{get, web, HttpResponse, Responder};
use actix_web_flash_messages::IncomingFlashMessages;
use serde::Deserialize;
use sqlx::{query_as, types::chrono::NaiveDate, PgPool};
use std::fmt::Write;
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
) -> impl Responder {
    let mut error_html = String::new();
    match writeln!(error_html, r#"<section class="container">"#) {
        Ok(_) => {}
        Err(e) => {
            return HttpResponse::InternalServerError().body(format!("{}", e));
        }
    }
    for m in flash_messages.iter() {
        match writeln!(error_html, r#"<p><i><mark>{}</mark></i></p>"#, m.content()) {
            Ok(_) => {}
            Err(e) => {
                return HttpResponse::InternalServerError().body(format!("{}", e));
            }
        };
    }
    match writeln!(error_html, r#"</section>"#) {
        Ok(_) => {}
        Err(e) => {
            return HttpResponse::InternalServerError().body(format!("{}", e));
        }
    };
    let scores = match query.matchup_id {
        Some(matchup_id) => match match_summary(matchup_id, &pg_pool).await {
            Ok(res) => res,
            Err(e) => return see_other("/scores", Some(e)),
        },
        None => "".to_owned(),
    };
    let main_div = include_str!("get.html");
    let main_div = format!(
        "{}\n<main class=\"container\">{}{}</main>",
        &error_html, &main_div, &scores
    );
    let html = compose_html(&main_div);
    HttpResponse::Ok().body(html)
}

async fn match_summary(match_id: Uuid, pg_pool: &PgPool) -> Result<String, anyhow::Error> {
    get_match_information(match_id, &pg_pool).await?;
    let match_scores = get_match_scores(match_id, &pg_pool).await?;
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
