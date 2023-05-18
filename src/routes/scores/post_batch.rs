use actix_web::{
    post,
    web::{Data, Form},
    Responder,
};
use anyhow::Context;
use sqlx::PgPool;
use uuid::Uuid;

use crate::routes::routing_utils::see_other;
use crate::routes::scores::post::get_match_information;

use super::post::{save_match_score, MatchScoreForm};

#[derive(serde::Deserialize)]
pub struct FormData {
    matchup_id: Uuid,
    raw_matches_list: String,
}

#[post("/scores_batch")]
pub async fn save_scores_batch(form_data: Form<FormData>, pg_pool: Data<PgPool>) -> impl Responder {
    let _match_info = match get_match_information(form_data.matchup_id, &pg_pool).await {
        Ok(res) => res,
        Err(e) => {
            return see_other("/scores", Some(e));
        }
    };
    dbg!(&form_data.raw_matches_list);
    let match_scores = match parse_match_scores(form_data.matchup_id, &form_data.raw_matches_list) {
        Ok(res) => res,
        Err(e) => {
            return see_other(
                &format!("/scores?matchup_id={}", form_data.matchup_id),
                Some(e),
            )
        }
    };
    let e = save_scores(&pg_pool, form_data.matchup_id, match_scores)
        .await
        .err();
    see_other(&format!("/scores?matchup_id={}", form_data.matchup_id), e)
}

pub async fn save_scores(
    pg_pool: &PgPool,
    matchup_id: Uuid,
    scores: Vec<MatchScoreForm>,
) -> Result<(), anyhow::Error> {
    for score in scores {
        save_match_score(
            pg_pool,
            matchup_id,
            &score.winner_initials,
            score.score,
            &score.played_at,
        )
        .await?;
    }
    Ok(())
}

fn parse_match_scores(
    matchup_id: Uuid,
    raw_match_scores: &str,
) -> Result<Vec<MatchScoreForm>, anyhow::Error> {
    let elements: Vec<&str> = raw_match_scores.split("\r\n").collect();
    let mut parsed_elements = Vec::new();
    for element in elements {
        parsed_elements
            .push(MatchScoreForm::new(matchup_id, element).context("Failed to parse match score")?);
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
        let raw_match_scores = "2022-02-22 P1 16:1\n2022-02-22 P1 16:1";
        let matchup_id = Uuid::new_v4();
        let res = parse_match_scores(matchup_id, raw_match_scores).unwrap();
        assert!(res[0].matchup_id == matchup_id);
        assert!(res[1].matchup_id == matchup_id);
    }
}
