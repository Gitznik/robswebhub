use actix_web::{post, HttpResponse, Responder, http::header::LOCATION};
use actix_web_flash_messages::FlashMessage;

#[post("/scores")]
async fn save_scores() -> impl Responder {
    FlashMessage::info("Sorry, I have not implemented saving scores yet.").send();
    HttpResponse::SeeOther().insert_header((LOCATION, "/scores")).finish()
}
