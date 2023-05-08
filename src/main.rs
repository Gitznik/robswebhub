use actix_files::Files;
use actix_web::cookie::Key;
use actix_web::{web, App, HttpServer};
use actix_web_flash_messages::{storage::CookieMessageStore, FlashMessagesFramework};
use robswebhub::configuration::DatabaseSettings;
use robswebhub::routes::scores::matches::get::match_summary;
use robswebhub::{
    configuration::get_configuration,
    routes::{
        about::get::about,
        root::get::root,
        scores::{get::add_scores, post::save_scores},
    },
};
use sqlx::postgres::PgPoolOptions;
use sqlx::PgPool;

async fn get_connection_pool(configuration: &DatabaseSettings) -> PgPool {
    PgPoolOptions::new()
        .acquire_timeout(std::time::Duration::from_secs(2))
        .connect_lazy(&configuration.connection_string)
        .expect("Could not connect to the database")
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let configuration = get_configuration().unwrap();
    let message_store = CookieMessageStore::builder(Key::generate()).build();
    let message_framework = FlashMessagesFramework::builder(message_store).build();
    let pg_pool = get_connection_pool(&configuration.database).await;
    let pg_pool = web::Data::new(pg_pool);
    HttpServer::new(move || {
        App::new()
            .wrap(message_framework.clone())
            .service(root)
            .service(about)
            .service(add_scores)
            .service(save_scores)
            .service(Files::new("/images", "./images"))
            .service(match_summary)
            .app_data(pg_pool.clone())
    })
    .bind((
        configuration.application.host,
        configuration.application.port,
    ))?
    .run()
    .await
}
