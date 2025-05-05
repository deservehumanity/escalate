use std::{arch::x86_64::_MM_ROUND_NEAREST, fs::File, io::Read};

use bitcoin::*;
use wallet::Wallet;

mod wallet;

use clap::{Parser, Subcommand};

#[derive(Parser)]
#[command(name = "escalate", version, about = "A bitcoin wallet", long_about = None)]
struct Cli {
    #[command(subcommand)]
    command: Commands
}

#[derive(Subcommand)]
enum Commands {
    New,
    Ls,
    Ca,
    Receive,
    Send {
        #[arg(short, long)]
        to: String,

        #[arg(short, long)]
        amt: f64
    }
}

fn main() {
    let cli = Cli::parse();

    match &cli.command {
        Commands::New => {
            let wl = Wallet::new(Network::Bitcoin);
            
            println!("Please write down following phrase somewhere safe.");
            println!("Without it you will lose ability to recover access to your wallet. The order matters.");
            
            println!("-------------");    

            for word in wl.mnemonic.words() {
                print!("{word} ");
            }
            
            println!("\n-------------");
            wl.save("wallet.esc").expect("W103 Failed to save the wallet");
        }
        
        Commands::Ls => {
            let wl = Wallet::load("wallet.esc");
            for (domain, key) in &wl.derived_keys {
                println!("{domain} -> {}", key.address);
            }
        }
        
        Commands::Ca => {
            let wl = Wallet::load("wallet.esc");
            let latest_addr_id = if wl.next_address > 0 {
                wl.next_address - 1
            } else {
                0
            };

            let path = format!("m/84'/0'/0'/0/{}", latest_addr_id);

            if let Some(dk) = wl.derived_keys.get(&path) {
                println!("Addr {} -> {}", latest_addr_id, dk.address);
            } else {
                println!("No address found for index {}", latest_addr_id);
            }
        }

        Commands::Receive => {
            let mut wl = Wallet::load("wallet.esc");

            let dk = wl.get_new_address().unwrap();

            println!("Send BTC to {}", dk.address);

            wl.save("wallet.esc").expect("W103 Failed to save the wallet");
        }

        Commands::Send { to, amt } => {
            println!("{to}, {amt}")
        }
    }
}

