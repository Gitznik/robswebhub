use std::{
    collections::{BinaryHeap, HashSet},
    num::ParseIntError,
};

use actix_web::{
    post,
    web::{Data, Form},
    HttpResponse, Result,
};
use anyhow::{anyhow, Context};
use sqlx::{query_as, types::chrono::NaiveDate, PgPool};
use uuid::Uuid;

use crate::routes::routing_utils::{see_other, see_other_error};

use super::post_batch::{save_scores_to_db, MatchScores};

#[derive(Clone)]
pub struct Score {
    pub winner_score: i16,
    pub loser_score: i16,
}

#[derive(Clone)]
pub struct MatchScoreInput {
    pub matchup_id: Uuid,
    pub winner_initials: String,
    pub score: Score,
    pub played_at: NaiveDate,
}

impl MatchScoreInput {
    pub fn new_from_form(score_form: &MatchScoreForm) -> Result<Self, anyhow::Error> {
        let played_at = Self::parse_played_at(score_form.played_at.to_owned())
            .context("Failed to parse the played_at date")?;
        let winner_initials = score_form.winner_initials.to_owned();
        let matchup_id = score_form.matchup_id;
        let score =
            Self::parse_score(score_form.score.to_owned()).context("Failed to parse the score")?;
        Ok(Self {
            matchup_id,
            played_at,
            winner_initials,
            score,
        })
    }

    pub fn new_from_str(matchup_id: Uuid, raw_score: &str) -> Result<Self, anyhow::Error> {
        let elements: Vec<&str> = raw_score.split(' ').collect();
        let played_at = Self::parse_played_at(elements[0].to_owned())
            .context("Failed to parse the played_at date")?;
        let winner_initials = elements[1].to_owned();
        let score =
            Self::parse_score(elements[2].to_owned()).context("Failed to parse the score")?;
        Ok(Self {
            matchup_id,
            played_at,
            winner_initials,
            score,
        })
    }

    fn parse_played_at(raw_played_at: String) -> Result<NaiveDate, anyhow::Error> {
        Ok(NaiveDate::parse_from_str(&raw_played_at, "%Y-%m-%d")?)
    }

    fn parse_score(raw_score: String) -> Result<Score, anyhow::Error> {
        let parsed_scores: Result<Vec<i16>, ParseIntError> =
            raw_score.split(':').map(|s| s.parse::<i16>()).collect();
        let mut scores = BinaryHeap::from(parsed_scores?);
        if scores.len() != 2 {
            return Err(anyhow!("Score length does not match. Expected 2"));
        }
        Ok(Score {
            winner_score: scores.pop().unwrap(),
            loser_score: scores.pop().unwrap(),
        })
    }
}

#[derive(serde::Deserialize, Debug)]
pub struct MatchScoreForm {
    pub matchup_id: Uuid,
    pub winner_initials: String,
    pub score: String,
    pub played_at: String,
}

#[post("/scores_single")]
async fn save_scores(
    form_data: Form<MatchScoreForm>,
    pg_pool: Data<PgPool>,
) -> Result<HttpResponse> {
    let match_info = get_match_information(form_data.matchup_id, &pg_pool)
        .await
        .map_err(|e| see_other_error("/scores", Some(e)))?;
    let match_scores = MatchScoreInput::new_from_form(&form_data).map_err(|e| {
        see_other_error(
            &format!("/scores?matchup_id={}", form_data.matchup_id),
            Some(e),
        )
    })?;
    let match_scores = MatchScores::new(vec![match_scores], match_info).map_err(|e| {
        see_other_error(
            &format!("/scores?matchup_id={}", form_data.matchup_id),
            Some(e),
        )
    })?;
    match save_match_score(&pg_pool, match_scores).await {
        Ok(_) => Ok(see_other(
            &format!("/scores?matchup_id={}", form_data.matchup_id),
            None,
        )),
        Err(e) => Ok(see_other("/scores", Some(e))),
    }
}

#[derive(Debug)]
#[allow(dead_code)]
pub struct MatchInfo {
    pub id: Uuid,
    pub player_1: String,
    pub player_2: String,
}

impl MatchInfo {
    pub fn player_in_match(&self, player: &String) -> bool {
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

pub async fn save_match_score(
    pg_pool: &PgPool,
    match_scores: MatchScores,
) -> Result<(), anyhow::Error> {
    save_scores_to_db(pg_pool, match_scores).await?;
    Ok(())
}
