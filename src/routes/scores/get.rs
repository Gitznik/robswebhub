use actix_web::{get, HttpResponse, Responder};

#[get("/scores")]
async fn add_scores() -> impl Responder {
    HttpResponse::Ok().body(include_str!("get.html"))
}
