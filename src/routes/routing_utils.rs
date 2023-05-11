use actix_web::{http::header::LOCATION, HttpResponse};
use actix_web_flash_messages::FlashMessage;

pub fn see_other(location: &str, error: Option<anyhow::Error>) -> HttpResponse {
    if let Some(e) = error {
        FlashMessage::info(e.to_string()).send();
    }
    return HttpResponse::SeeOther()
        .insert_header((LOCATION, location))
        .finish();
}
