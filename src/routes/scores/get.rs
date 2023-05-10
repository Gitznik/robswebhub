use actix_web::{get, HttpResponse, Responder};
use actix_web_flash_messages::IncomingFlashMessages;
use std::fmt::Write;

use crate::html_base::compose_html;

#[get("/scores")]
async fn add_scores(flash_messages: IncomingFlashMessages) -> impl Responder {
    let mut error_html = String::new();
    match writeln!(error_html, r#"<section class="container">"#) {
        Ok(_) => {}
        Err(e) => {
            return HttpResponse::InternalServerError().body(format!("{}", e));
        }
    }
    for m in flash_messages.iter() {
        match writeln!(error_html, r#"<p><i><mark>{}</mark></i></p>"#, m.content()) {
            Ok(_) => {}
            Err(e) => {
                return HttpResponse::InternalServerError().body(format!("{}", e));
            }
        };
    }
    match writeln!(error_html, r#"</section>"#) {
        Ok(_) => {}
        Err(e) => {
            return HttpResponse::InternalServerError().body(format!("{}", e));
        }
    };
    let main_div = include_str!("get.html");
    let main_div = format!("{}\n{}", &error_html, &main_div);
    let html = compose_html(&main_div);
    HttpResponse::Ok().body(html)
}
