use crate::{html_base::compose_html, routes::scores::post::get_match_information};
use actix_web::{get, http::header::LOCATION, web, HttpResponse, Responder};
use actix_web_flash_messages::FlashMessage;
use sqlx::{
    query_as,
    types::chrono::{self, Utc},
    PgPool,
};
use uuid::Uuid;

#[get("/scores/{match_id}")]
async fn match_summary(path: web::Path<(Uuid,)>, pg_pool: web::Data<PgPool>) -> impl Responder {
    let match_id = path.into_inner().0;
    match get_match_information(match_id, &pg_pool).await {
        Ok(res) => res,
        Err(e) => {
            FlashMessage::info(e.to_string()).send();
            return HttpResponse::SeeOther()
                .insert_header((LOCATION, "/scores"))
                .finish();
        }
    };
    let match_scores = match get_match_scores(match_id, &pg_pool).await {
        Ok(res) => res,
        Err(e) => {
            FlashMessage::info(e.to_string()).send();
            return HttpResponse::SeeOther()
                .insert_header((LOCATION, "/scores"))
                .finish();
        }
    };

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
                res.winner, res.winner_score, res.loser_score, res.created_at
            )
        })
        .collect();

    let html = format!(
        r#"
      <main class="container">
        <h1>Match Scores for Match {}</h1>
        <table role="grid">
          <thead>
            <tr>
              <th scope="col">#</th>
              <th scope="col">winner</th>
              <th scope="col">winner_score</th>
              <th scope="col">loser_score</th>
              <th scope="col">created_at</th>
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
    );
    let html = compose_html(&html);

    HttpResponse::Ok().body(html)
}

#[derive(Debug)]
#[allow(dead_code)]
struct MatchScore {
    match_id: Uuid,
    game_id: Uuid,
    winner: String,
    created_at: chrono::DateTime<Utc>,
    winner_score: i16,
    loser_score: i16,
}

async fn get_match_scores(
    matchup_id: Uuid,
    pg_pool: &PgPool,
) -> Result<Vec<MatchScore>, sqlx::Error> {
    query_as!(
        MatchScore,
        r#"
        select match_id, game_id, winner, created_at, winner_score, loser_score
        from scores
        where match_id = $1
        "#,
        matchup_id
    )
    .fetch_all(pg_pool)
    .await
}
