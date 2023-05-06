use actix_web::{post, HttpResponse, Responder, http::header::LOCATION};

#[post("/scores")]
async fn save_scores() -> impl Responder {
    HttpResponse::SeeOther().insert_header((LOCATION, "/scores")).finish()
}
