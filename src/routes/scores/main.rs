use actix_web::web;

use super::{
    components::{
        batch_upload_form::get_batch_result_form, single_result_form::get_single_result_form,
    },
    get::add_scores,
    post::save_scores,
    post_batch::save_scores_batch,
};

pub fn scores_config(cfg: &mut web::ServiceConfig) {
    cfg.service(
        web::scope("/scores")
            .service(save_scores)
            .service(add_scores)
            .service(save_scores_batch)
            .service(get_single_result_form)
            .service(get_batch_result_form),
    );
}
