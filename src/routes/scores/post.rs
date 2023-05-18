use std::collections::{BinaryHeap, HashSet};

use actix_web::{
    post,
    web::{Data, Form},
    Responder,
};
use sqlx::{query, query_as, types::chrono::NaiveDate, PgPool};
use uuid::Uuid;

use crate::routes::routing_utils::see_other;

#[derive(serde::Deserialize, Debug)]
pub struct MatchScoreForm {
    pub matchup_id: Uuid,
    pub winner_initials: String,
    pub score: String,
    pub played_at: String,
}

#[post("/scores")]
async fn save_scores(form_data: Form<MatchScoreForm>, pg_pool: Data<PgPool>) -> impl Responder {
    let match_info = match get_match_information(form_data.matchup_id, &pg_pool).await {
        Ok(res) => res,
        Err(e) => {
            return see_other("/scores", Some(e));
        }
    };
    if match_info.player_in_match(&form_data.winner_initials) {
        match save_match_score(
            &pg_pool,
            form_data.matchup_id,
            &form_data.winner_initials,
            form_data.score.clone(),
            &form_data.played_at,
        )
        .await
        {
            Ok(_) => see_other(
                &format!("/scores?matchup_id={}", form_data.matchup_id),
                None,
            ),
            Err(e) => see_other("/scores", Some(e)),
        }
    } else {
        see_other(
            &format!("/scores?matchup_id={}", form_data.matchup_id),
            Some(anyhow::anyhow!("Player not found in this match")),
        )
    }
}

#[derive(Debug)]
#[allow(dead_code)]
pub struct MatchInfo {
    id: Uuid,
    player_1: String,
    player_2: String,
}

impl MatchInfo {
    fn player_in_match(&self, player: &String) -> bool {
        let players_in_match = HashSet::from([&self.player_1, &self.player_2]);
        players_in_match.contains(&player)
    }
}

pub async fn get_match_information(
    matchup_id: Uuid,
    pg_pool: &PgPool,
) -> Result<MatchInfo, anyhow::Error> {
    let match_info = query_as!(
        MatchInfo,
        r#"
        select id, player_1, player_2
        from matches
        where id = $1
        "#,
        matchup_id
    )
    .fetch_one(pg_pool)
    .await?;
    Ok(match_info)
}

async fn save_match_score(
    pg_pool: &PgPool,
    matchup_id: Uuid,
    winner_initials: &str,
    score: String,
    played_date: &str,
) -> Result<(), anyhow::Error> {
    let mut scores = BinaryHeap::from(
        score
            .split(':')
            .map(|s| s.parse::<i16>().unwrap())
            .collect::<Vec<i16>>(),
    );
    let game_id = Uuid::new_v4();
    let parsed_date = NaiveDate::parse_from_str(played_date, "%Y-%m-%d")?;
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
    Ok(())
}
