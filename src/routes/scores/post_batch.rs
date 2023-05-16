use actix_web::{
    post,
    web::{Data, Form},
    Responder,
};
use sqlx::PgPool;
use uuid::Uuid;

use crate::routes::routing_utils::see_other;
use crate::routes::scores::post::get_match_information;

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
