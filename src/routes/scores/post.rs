use std::collections::BinaryHeap;

use actix_web::{
    http::header::LOCATION,
    post,
    web::{Data, Form},
    HttpResponse, Responder,
};
use actix_web_flash_messages::FlashMessage;
use sqlx::{query, query_as, types::chrono::NaiveDate, PgPool};
use uuid::Uuid;

#[derive(serde::Deserialize)]
pub struct FormData {
    matchup_id: Uuid,
    winner_initials: String,
    score: String,
    played_at: String,
}

#[post("/scores")]
async fn save_scores(form_data: Form<FormData>, pg_pool: Data<PgPool>) -> impl Responder {
    match get_match_information(form_data.matchup_id, &pg_pool).await {
        Ok(res) => res,
        Err(e) => {
            FlashMessage::info(e.to_string()).send();
            return HttpResponse::SeeOther()
                .insert_header((LOCATION, "/scores"))
                .finish();
        }
    };
    save_match_score(
        &pg_pool,
        form_data.matchup_id,
        &form_data.winner_initials,
        form_data.score.clone(),
        &form_data.played_at,
    )
    .await;
    HttpResponse::SeeOther()
        .insert_header((LOCATION, format!("/scores/{}", form_data.matchup_id)))
        .finish()
}

#[derive(Debug)]
#[allow(dead_code)]
pub struct MatchInfo {
    id: Uuid,
    player_1: String,
    player_2: String,
}

pub async fn get_match_information(
    matchup_id: Uuid,
    pg_pool: &PgPool,
) -> Result<MatchInfo, sqlx::Error> {
    query_as!(
        MatchInfo,
        r#"
        select id, player_1, player_2
        from matches
        where id = $1
        "#,
        matchup_id
    )
    .fetch_one(pg_pool)
    .await
}

async fn save_match_score(
    pg_pool: &PgPool,
    matchup_id: Uuid,
    winner_initials: &str,
    score: String,
    played_date: &str,
) {
    let mut scores = BinaryHeap::from(
        score
            .split(':')
            .map(|s| s.parse::<i16>().unwrap())
            .collect::<Vec<i16>>(),
    );
    let game_id = Uuid::new_v4();
    let parsed_date = NaiveDate::parse_from_str(played_date, "%Y-%m-%d")
        .unwrap_or_else(|_| panic!("Could not parse date from form: {}", played_date));
    query!(
        r#"
        INSERT INTO scores (match_id, game_id, winner, winner_score, loser_score, created_at, played_at)
        VALUES ($1, $2, $3, $4, $5, now(), $6)
        "#,
        matchup_id,
        game_id,
        winner_initials,
        scores.pop(),
        scores.pop(),
        parsed_date,
    )
    .execute(pg_pool)
    .await
    .expect("Could not save the score");
}
