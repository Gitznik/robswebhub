use std::{collections::HashSet, ops::Deref};

use actix_web::{
    post,
    web::{Data, Form},
    Responder,
};
use anyhow::{anyhow, Context};
use sqlx::{
    types::chrono::{NaiveDate, NaiveDateTime, Utc},
    PgPool,
};
use uuid::Uuid;

use crate::routes::routing_utils::see_other;
use crate::routes::scores::post::get_match_information;

use super::post::{MatchInfo, MatchScoreInput};

#[derive(serde::Deserialize)]
pub struct FormData {
    matchup_id: Uuid,
    raw_matches_list: String,
}

#[post("/scores_batch")]
pub async fn save_scores_batch(form_data: Form<FormData>, pg_pool: Data<PgPool>) -> impl Responder {
    let match_info = match get_match_information(form_data.matchup_id, &pg_pool).await {
        Ok(res) => res,
        Err(e) => {
            return see_other("/scores", Some(e));
        }
    };
    let match_scores = match parse_match_scores(form_data.matchup_id, &form_data.raw_matches_list) {
        Ok(res) => res,
        Err(e) => {
            return see_other(
                &format!("/scores?matchup_id={}", form_data.matchup_id),
                Some(e),
            )
        }
    };
    let match_scores = match MatchScores::new(match_scores, match_info) {
        Ok(res) => res,
        Err(e) => {
            return see_other(
                &format!("/scores?matchup_id={}", form_data.matchup_id),
                Some(e),
            )
        }
    };
    let e = save_scores_to_db(&pg_pool, match_scores).await.err();
    see_other(&format!("/scores?matchup_id={}", form_data.matchup_id), e)
}

pub struct MatchScores(Vec<MatchScoreInput>);

impl MatchScores {
    pub fn new(
        match_scores: Vec<MatchScoreInput>,
        match_info: MatchInfo,
    ) -> Result<Self, anyhow::Error> {
        let players: HashSet<String> = match_scores
            .clone()
            .into_iter()
            .map(|score| score.winner_initials)
            .collect();

        for player in players {
            if !match_info.player_in_match(&player) {
                return Err(anyhow!("Provided player not in match"));
            }
        }

        Ok(MatchScores(match_scores))
    }
}

impl IntoIterator for MatchScores {
    type Item = MatchScoreInput;
    type IntoIter = <Vec<MatchScoreInput> as IntoIterator>::IntoIter;

    fn into_iter(self) -> Self::IntoIter {
        self.0.into_iter()
    }
}

impl Deref for MatchScores {
    type Target = Vec<MatchScoreInput>;

    fn deref(&self) -> &Self::Target {
        &self.0
    }
}

pub async fn save_scores_to_db(pg_pool: &PgPool, scores: MatchScores) -> Result<(), anyhow::Error> {
    let mut matchup_ids: Vec<Uuid> = Vec::with_capacity(scores.len());
    let mut winner_initials: Vec<String> = Vec::with_capacity(scores.len());
    let mut winner_scores: Vec<i16> = Vec::with_capacity(scores.len());
    let mut loser_scores: Vec<i16> = Vec::with_capacity(scores.len());
    let mut played_at: Vec<NaiveDate> = Vec::with_capacity(scores.len());
    let mut created_at: Vec<NaiveDateTime> = Vec::with_capacity(scores.len());
    let mut game_ids: Vec<Uuid> = Vec::with_capacity(scores.len());
    scores.into_iter().for_each(|row| {
        matchup_ids.push(row.matchup_id);
        winner_initials.push(row.winner_initials);
        winner_scores.push(row.score.winner_score);
        loser_scores.push(row.score.loser_score);
        played_at.push(row.played_at);
        created_at.push(Utc::now().naive_utc());
        game_ids.push(Uuid::new_v4());
    });
    sqlx::query(r#"
        INSERT INTO scores (match_id, game_id, winner, winner_score, loser_score, created_at, played_at)
        SELECT * FROM UNNEST ($1,$2,$3,$4,$5,$6,$7)"#)
    .bind(matchup_ids)
    .bind(game_ids)
    .bind(winner_initials)
    .bind(winner_scores)
    .bind(loser_scores)
    .bind(created_at)
    .bind(played_at)
    .execute(pg_pool)
    .await?;
    Ok(())
}

fn parse_match_scores(
    matchup_id: Uuid,
    raw_match_scores: &str,
) -> Result<Vec<MatchScoreInput>, anyhow::Error> {
    let elements: Vec<&str> = raw_match_scores.split("\r\n").collect();
    let mut parsed_elements = Vec::new();
    for element in elements {
        parsed_elements.push(
            MatchScoreInput::new_from_str(matchup_id, element)
                .context("Failed to parse match score")?,
        );
    }
    Ok(parsed_elements)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn parsing_single_score_works() {
        let raw_match_scores = r#"2022-02-22 P1 16:1"#;
        let matchup_id = Uuid::new_v4();
        let res = parse_match_scores(matchup_id, raw_match_scores).unwrap();
        assert!(res[0].matchup_id == matchup_id);
    }

    #[test]
    fn parsing_multiple_scores_works() {
        let raw_match_scores = "2022-02-22 P1 16:1\r\n2022-02-22 P1 16:1";
        let matchup_id = Uuid::new_v4();
        let res = parse_match_scores(matchup_id, raw_match_scores).unwrap();
        assert!(res[0].matchup_id == matchup_id);
        assert!(res[1].matchup_id == matchup_id);
    }
}
