use crate::routes::{
    flash_messages_utils::flash_messages_section, routing_utils::see_other_error,
    scores::post::get_match_information,
};
use actix_web::{get, web, HttpResponse};
use actix_web_flash_messages::IncomingFlashMessages;
use chrono::Days;
use core::ops::Deref;
use itertools::Itertools;
use plotters::prelude::*;
use serde::Deserialize;
use sqlx::{query_as, types::chrono::NaiveDate, PgPool};
use std::collections::HashMap;
use uuid::Uuid;

use crate::html_base::compose_html;

use super::post::MatchInfo;

#[derive(Debug, Deserialize)]
struct QueryData {
    matchup_id: Option<Uuid>,
}

#[derive(Debug, Clone)]
struct MatchScores(Vec<MatchScore>);

impl MatchScores {
    fn players(&self) -> Vec<String> {
        self.clone()
            .into_iter()
            .unique_by(|s| s.clone().winner)
            .map(|s| s.winner)
            .collect_vec()
    }

    fn cumm_sum_wins(&self) -> HashMap<String, Vec<(NaiveDate, i32)>> {
        let players = self.players();
        let mut scores = HashMap::new();
        for player in players {
            let cum_scores = self
                .clone()
                .into_iter()
                .filter(|s| s.winner == player)
                .map(|s| s.played_at)
                .sorted()
                .scan(0, |acc, d| {
                    *acc += 1;
                    Some((d, *acc))
                })
                .collect_vec();

            scores.insert(player, cum_scores);
        }
        scores
    }
}

impl IntoIterator for MatchScores {
    type Item = MatchScore;
    type IntoIter = <Vec<MatchScore> as IntoIterator>::IntoIter;

    fn into_iter(self) -> Self::IntoIter {
        self.0.into_iter()
    }
}

impl Deref for MatchScores {
    type Target = Vec<MatchScore>;

    fn deref(&self) -> &Self::Target {
        &self.0
    }
}

#[get("")]
async fn add_scores(
    query: web::Query<QueryData>,
    pg_pool: web::Data<PgPool>,
    flash_messages: IncomingFlashMessages,
) -> actix_web::Result<HttpResponse> {
    let error_html = flash_messages_section(flash_messages)
        .map_err(actix_web::error::ErrorInternalServerError)?;
    let scores = match query.matchup_id {
        Some(matchup_id) => match_summary(matchup_id, &pg_pool)
            .await
            .map_err(|e| see_other_error("/scores", Some(e)))?,
        None => "".to_owned(),
    };
    let insert_score_form = insert_score_form(query.matchup_id);
    let main_div = include_str!("get.html");
    let main_div = format!(
        "{}\n<main class=\"container\">{}{}{}</main>",
        &error_html, &main_div, &insert_score_form, &scores
    );
    let html = compose_html(&main_div);
    Ok(HttpResponse::Ok().body(html))
}

async fn match_summary(match_id: Uuid, pg_pool: &PgPool) -> Result<String, anyhow::Error> {
    let match_information = get_match_information(match_id, pg_pool).await?;
    let match_scores = MatchScores(get_match_scores(match_id, pg_pool).await?);

    match_result_plots(match_information, match_scores.clone()).unwrap();

    let match_rows: Vec<String> = match_scores
        .into_iter()
        .map(|res| {
            format!(
                r#"
            <tr>
              <th scope="row">1</th>
              <td>{}</td>
              <td>{}</td>
              <td>{}</td>
              <td>{}</td>
            </tr>
        "#,
                res.winner, res.winner_score, res.loser_score, res.played_at
            )
        })
        .collect();

    Ok(format!(
        r#"
        <h1>Match Scores for Match</h1>
        <h2>{}</h2>
        <table role="grid">
          <thead>
            <tr>
              <th scope="col">#</th>
              <th scope="col">winner</th>
              <th scope="col">winner_score</th>
              <th scope="col">loser_score</th>
              <th scope="col">played_at</th>
            </tr>
          </thead>
          <tbody>
            {}
          </tbody>
        </table>
        <div>
          <img src="images/match_plots/{}.png" alt="Match Results Graph" width="640" height="480">
        </div>
      </main>
    "#,
        &match_id,
        match_rows.join("\n"),
        &match_id,
    ))
}

fn match_result_plots(
    match_information: MatchInfo,
    match_scores: MatchScores,
) -> Result<(), anyhow::Error> {
    let mut wins = match_scores.cumm_sum_wins();
    let p1_wins = wins.remove(&match_information.player_1).unwrap_or_else(|| {
        let v: Vec<(NaiveDate, i32)> = Vec::new();
        v
    });
    let p2_wins = wins.remove(&match_information.player_2).unwrap_or_else(|| {
        let v: Vec<(NaiveDate, i32)> = Vec::new();
        v
    });
    let max_wins = std::cmp::max(
        p1_wins
            .clone()
            .into_iter()
            .max_by_key(|w| w.1)
            .map(|w| w.1)
            .unwrap_or(0),
        p2_wins
            .clone()
            .into_iter()
            .max_by_key(|w| w.1)
            .map(|w| w.1)
            .unwrap_or(0),
    );

    let path = format!("images/match_plots/{}.png", match_information.id);
    let root = BitMapBackend::new(&path, (640, 480)).into_drawing_area();
    let (start, end) = match_scores
        .into_iter()
        .minmax_by_key(|s| s.played_at)
        .into_option()
        .unwrap();
    root.fill(&WHITE)?;
    let mut chart = ChartBuilder::on(&root)
        .caption("Summary of Wins", ("sans-serif", 50).into_font())
        .margin(5)
        .x_label_area_size(30)
        .y_label_area_size(30)
        .build_cartesian_2d(
            start.played_at.checked_sub_days(Days::new(9)).unwrap()
                ..end.played_at.checked_add_days(Days::new(9)).unwrap(),
            0..(max_wins + (max_wins / 20 + 1)),
        )?;

    chart.configure_mesh().draw()?;
    const STROKE_WIDTH: u32 = 3;
    chart
        .draw_series(LineSeries::new(p1_wins, BLUE.stroke_width(STROKE_WIDTH)))?
        .label(format!("Wins of {}", &match_information.player_1))
        .legend(|(x, y)| {
            PathElement::new(vec![(x, y), (x + 20, y)], BLUE.stroke_width(STROKE_WIDTH))
        });
    chart
        .draw_series(LineSeries::new(p2_wins, RED.stroke_width(STROKE_WIDTH)))?
        .label(format!("Wins of {}", &match_information.player_2))
        .legend(|(x, y)| {
            PathElement::new(vec![(x, y), (x + 20, y)], RED.stroke_width(STROKE_WIDTH))
        });

    chart
        .configure_series_labels()
        .background_style(WHITE.mix(0.8))
        .border_style(BLACK)
        .draw()?;

    root.present()?;

    Ok(())
}

#[derive(Debug, Clone)]
#[allow(dead_code)]
struct MatchScore {
    match_id: Uuid,
    game_id: Uuid,
    winner: String,
    played_at: NaiveDate,
    winner_score: i16,
    loser_score: i16,
}

async fn get_match_scores(
    matchup_id: Uuid,
    pg_pool: &PgPool,
) -> Result<Vec<MatchScore>, anyhow::Error> {
    let scores = query_as!(
        MatchScore,
        r#"
        select match_id, game_id, winner, played_at, winner_score, loser_score
        from scores
        where match_id = $1
        order by played_at desc
        "#,
        matchup_id
    )
    .fetch_all(pg_pool)
    .await?;
    Ok(scores)
}

fn insert_score_form(matchup_id: Option<Uuid>) -> String {
    match matchup_id {
        Some(matchup_id) => {
            format!(
                r##"
        <div>
          <h2>Add Score</h2>
          <div class="grid">
            <div class="grid">
              <button hx-get="/scores/single_result_form" hx-trigger="load,click" hx-target="#score_entry_form" hx-vals='{{"matchup_id":"{}"}}' hx-swap="outerHTML ignoreTitle:true">Single Result Entry</button>
              <button hx-get="/scores/batch_result_form" hx-target="#score_entry_form" hx-vals='{{"matchup_id":"{}"}}' hx-swap="outerHTML ignoreTitle:true">Multiple Result Entry</button>
            </div>
          </div>
          <div id="score_entry_form"</div>
        </div>
    "##,
                matchup_id, matchup_id
            )
        }
        None => "".to_owned(),
    }
}
