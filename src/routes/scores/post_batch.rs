use actix_web::{
    post,
    web::{Data, Form},
    Responder,
};
use sqlx::PgPool;
use uuid::Uuid;

use crate::routes::routing_utils::see_other;
use crate::routes::scores::post::get_match_information;

use super::post::MatchScoreForm;

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
    see_other(
        &format!("/scores?matchup_id={}", form_data.matchup_id),
        Some(anyhow::anyhow!("Saving batches not yet implemented")),
    )
}

fn parse_match_scores(matchup_id: Uuid, raw_match_scores: &str) -> Vec<MatchScoreForm> {
    let elements: Vec<&str> = raw_match_scores.split("\n").collect();
    let mut parsed_elements = Vec::new();
    for element in elements {
        parsed_elements.push(parse_map_score(matchup_id, element));
    }
    parsed_elements
}

fn parse_map_score(matchup_id: Uuid, raw_match_score: &str) -> MatchScoreForm {
    let elements: Vec<&str> = raw_match_score.split(' ').collect();
    MatchScoreForm {
        matchup_id,
        played_at: elements[0].to_owned(),
        winner_initials: elements[1].to_owned(),
        score: elements[2].to_owned(),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn parsing_single_score_works() {
        let raw_match_scores = r#"2022-02-22 P1 16:1"#;
        let matchup_id = Uuid::new_v4();
        let res = parse_match_scores(matchup_id, raw_match_scores);
        assert!(res[0].matchup_id == matchup_id);
    }

    #[test]
    fn parsing_multiple_scores_works() {
        let raw_match_scores = "2022-02-22 P1 16:1\n2022-02-22 P1 16:1";
        let matchup_id = Uuid::new_v4();
        let res = parse_match_scores(matchup_id, raw_match_scores);
        assert!(res[0].matchup_id == matchup_id);
        assert!(res[1].matchup_id == matchup_id);
    }
}
