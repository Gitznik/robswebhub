use std::fmt::Display;
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


pub fn see_other_error(location: &str, error: Option<anyhow::Error>) -> SeeOtherError {
    if let Some(e) = error {
        FlashMessage::info(e.to_string()).send();
    }
    SeeOtherError::SeeOther(location.to_owned())
}

#[derive(Debug)]
pub enum SeeOtherError {
    SeeOther(String),
}

impl Display for SeeOtherError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "Redirecting due to user facing error")
    }
}

impl actix_web::error::ResponseError for SeeOtherError {
    fn error_response(&self) -> HttpResponse<actix_web::body::BoxBody> {
        match self {
            SeeOtherError::SeeOther(location) => HttpResponse::SeeOther()
                .insert_header((LOCATION, location.to_owned()))
                .finish(),
        }
    }
}
