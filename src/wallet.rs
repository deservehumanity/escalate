use std::{collections::HashMap, fs::OpenOptions, io::Read, str::FromStr};
use bip39::{Language, Mnemonic};
use bitcoin::{bip32::{DerivationPath, Xpriv}, key::Secp256k1, Address, CompressedPublicKey, Network, PrivateKey, PublicKey};

use std::fs::File;
use std::io::{self, Write};

use serde::*;
use serde_json;

pub struct Wallet {
    pub mnemonic: Mnemonic,
    pub master_key: Xpriv,
    pub network: Network,
    pub derived_keys: HashMap<String, DerivedKey>,
    pub next_address: u128,
}

pub struct DerivedKey {
    pub path: String,
    pub private_key_wif: String,
    pub public_key: String,
    pub address: String
}

#[derive(Serialize, Deserialize, Debug)]
struct InnerWallet {
    mnemonic: Mnemonic,
    next_address: u128,
}

impl Wallet {
    pub fn new(network: Network) -> Self {
        let mnemonic = Mnemonic::generate_in(
            Language::English,
            12
        ).expect("W100 Mnemonic generation failed");

        let seed = mnemonic.to_seed("");

        let master_key = Xpriv::new_master(
            network,
            &seed
        ).expect("W101 Master key generation failed");
        
        
        Self {
            mnemonic,
            master_key,
            network,
            derived_keys: HashMap::new(),
            next_address: 0,
        }
    }
    
    pub fn save(&self, path: &str) -> io::Result<()> {
        let inner_wallet = InnerWallet {
            mnemonic: self.mnemonic.clone(),
            next_address: self.next_address
        };

        let serialized_wallet = serde_json::to_string(&inner_wallet).unwrap();
        
        let mut file = OpenOptions::new()
            .write(true)
            .truncate(true)
            .create(true)
            .open(path)?;
        
        file.write_all(serialized_wallet.as_bytes())?;

        Ok(())
    }
    
    pub fn load(path: &str) -> Self {
        let mut file = File::open(path).unwrap();
        let mut contents = String::new();
        
        file.read_to_string(&mut contents).unwrap();

        let wallet: InnerWallet = serde_json::from_str(&contents).unwrap();
        
        let master_key = Xpriv::new_master(
            Network::Bitcoin,
            &wallet.mnemonic.to_seed("")
        ).unwrap();
        

        let mut wl = Self {
            mnemonic: wallet.mnemonic,
            master_key,
            next_address: 0,
            derived_keys: HashMap::new(),
            network: Network::Bitcoin
        };
        
        for _ in 0..wallet.next_address {
            wl.get_new_address().unwrap();
        };

        wl
    }

    pub fn from_mnemonic(mnemonic: &str, network: Network) -> Self {
        let mnem = Mnemonic::from_str(mnemonic).unwrap();

        let master_key = Xpriv::new_master(
            network, 
            &mnem.to_seed("")
        ).expect("W101 Master key generation failed");

        Wallet {
            mnemonic: mnem,
            master_key,
            network,
            derived_keys: HashMap::new(),
            next_address: 0,
        }
    }

    pub fn derive_key(&mut self, path: &str) -> Option<&DerivedKey> {
        let secp = Secp256k1::new();

        let derivation_path = path.parse::<DerivationPath>().unwrap();

        let derived_xpriv = self.master_key.derive_priv(
            &secp, 
            &derivation_path
        ).unwrap();

        let private_key = PrivateKey {
            inner: derived_xpriv.private_key,
            compressed: true,
            network: self.network.into()
        };

        let public_key = CompressedPublicKey::from_private_key(&secp, &private_key).unwrap();

        let address = Address::p2wpkh(
            &public_key,
            self.network
        );

        let derived = DerivedKey {
            path: path.to_string(),
            private_key_wif: private_key.to_wif(),
            public_key: public_key.to_string(),
            address: address.to_string()
        };

        self.derived_keys.insert(path.to_string(), derived);
        
        self.next_address += 1;

        Some(self.derived_keys.get(path).unwrap())
    }

    pub fn get_new_address(&mut self) -> Option<&DerivedKey> {
        let index = self.next_address;

        let path = format!("m/84'/0'/0'/0/{}", index);
        self.derive_key(&path).unwrap();

        self.derived_keys.get(&path)
    }
}
