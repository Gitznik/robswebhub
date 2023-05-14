use actix_web_flash_messages::IncomingFlashMessages;
use std::fmt::Write;

pub fn flash_messages_section(flash_messages: IncomingFlashMessages) -> Result<String, anyhow::Error> {
    let mut error_html = String::new();
    writeln!(error_html, r#"<section class="container">"#)?;
    for m in flash_messages.iter() {
        writeln!(error_html, r#"<p><i><mark>{}</mark></i></p>"#, m.content())?
    }
    writeln!(error_html, r#"</section>"#)?;
    Ok(error_html)
}
