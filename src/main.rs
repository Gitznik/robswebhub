use actix_files::Files;
use actix_web::cookie::Key;
use actix_web::{App, HttpServer};
use actix_web_flash_messages::{storage::CookieMessageStore, FlashMessagesFramework};
use robswebhub::{
    configuration::get_configuration,
    routes::{
        about::get::about,
        root::get::root,
        scores::{get::add_scores, post::save_scores},
    },
};

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let config = get_configuration().unwrap();
    let message_store = CookieMessageStore::builder(Key::generate()).build();
    let message_framework = FlashMessagesFramework::builder(message_store).build();
    HttpServer::new(move || {
        App::new()
            .wrap(message_framework.clone())
            .service(root)
            .service(about)
            .service(add_scores)
            .service(save_scores)
            .service(Files::new("/images", "./images"))
    })
    .bind((config.application.host, config.application.port))?
    .run()
    .await
}
