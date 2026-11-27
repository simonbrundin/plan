//! Glob-mönster enligt Go:s `filepath.Glob`.
//!
//! Go stöder metatecknen `*`, `?` och `[...]`. `*` matchar tecken som
//! inte är en katalogseparator, så `**` betyder samma som `*` - mönstret
//! matchar alltså en katalognivå per `*`.

use std::path::{Path, PathBuf};

/// Om mönstret innehåller metatecken.
pub fn har_meta(mönster: &str) -> bool {
    mönster.contains('*') || mönster.contains('?') || mönster.contains('[')
}

/// Hittar filer som matchar mönstret.
///
/// Go: `filepath.Glob`. I/O-fel ignoreras - som i Go returneras då
/// inga träffar.
pub fn glob(mönster: &str) -> Result<Vec<PathBuf>, String> {
    // Ett mönster utan metatecken finns bara om filen finns.
    if !har_meta(mönster) {
        let sökväg = PathBuf::from(mönster);

        return Ok(if sökväg.exists() {
            vec![sökväg]
        } else {
            Vec::new()
        });
    }

    let absolut = mönster.starts_with('/');

    let segment: Vec<&str> = mönster.split('/').filter(|del| !del.is_empty()).collect();

    let mut nuvarande: Vec<PathBuf> = vec![if absolut {
        PathBuf::from("/")
    } else {
        PathBuf::new()
    }];

    for del in segment {
        let mut nästa: Vec<PathBuf> = Vec::new();

        if !har_meta(del) {
            nuvarande = nuvarande
                .into_iter()
                .map(|mapp| mapp.join(del))
                .filter(|sökväg| sökväg.exists())
                .collect();

            continue;
        }

        for mapp in nuvarande {
            let Ok(poster) = std::fs::read_dir(&mapp) else {
                continue;
            };

            let mut namn: Vec<String> = poster
                .filter_map(|post| post.ok())
                .map(|post| post.file_name().to_string_lossy().to_string())
                .collect();

            // Go läser katalogen i sorterad ordning.
            namn.sort();

            for namn in namn {
                if matchar_segment(del, &namn) {
                    nästa.push(mapp.join(&namn));
                }
            }
        }

        nuvarande = nästa;
    }

    Ok(nuvarande)
}

/// Matchar ett mönstersegment mot ett filnamn.
///
/// Go: `filepath.Match`.
pub fn matchar_segment(mönster: &str, namn: &str) -> bool {
    matchar(mönster.chars().collect(), namn.chars().collect())
}

/// Matchar mönstret mot namnet - port av Go:s `matchChunk`.
fn matchar(mönster: Vec<char>, namn: Vec<char>) -> bool {
    matchar_chunk(&mönster, &namn)
}

/// Port av Go:s `matchChunk`.
fn matchar_chunk(mönster: &[char], namn: &[char]) -> bool {
    let mut mönster = mönster;
    let mut namn = namn;
    let mut misslyckad = false;

    while !mönster.is_empty() {
        if !misslyckad && namn.is_empty() {
            misslyckad = true;
        }

        match mönster[0] {
            '[' => {
                let tecken = if misslyckad {
                    '\0'
                } else {
                    match namn.first() {
                        Some(tecken) => *tecken,
                        None => '\0',
                    }
                };

                if !misslyckad && !namn.is_empty() {
                    namn = &namn[1..];
                }

                mönster = &mönster[1..];

                let mut negativ = false;
                if !mönster.is_empty() && mönster[0] == '^' {
                    negativ = true;
                    mönster = &mönster[1..];
                }

                let mut träff = false;
                let mut antal = 0;

                loop {
                    if !mönster.is_empty() && mönster[0] == ']' && antal > 0 {
                        mönster = &mönster[1..];
                        break;
                    }

                    if mönster.is_empty() {
                        // Trasig klass - Go returnerar ett fel, här
                        // matchar mönstret inte.
                        return false;
                    }

                    let låg = mönster[0];
                    mönster = &mönster[1..];

                    let mut hög = låg;

                    if !mönster.is_empty() && mönster[0] == '-' {
                        mönster = &mönster[1..];

                        if mönster.is_empty() {
                            return false;
                        }

                        hög = mönster[0];
                        mönster = &mönster[1..];
                    }

                    if låg <= tecken && tecken <= hög {
                        träff = true;
                    }

                    antal += 1;
                }

                if träff == negativ {
                    misslyckad = true;
                }
            }
            '?' => {
                if !misslyckad {
                    if namn.is_empty() {
                        return false;
                    }

                    namn = &namn[1..];
                }

                mönster = &mönster[1..];
            }
            '*' => {
                mönster = &mönster[1..];

                while !mönster.is_empty() && mönster[0] == '*' {
                    mönster = &mönster[1..];
                }

                if mönster.is_empty() {
                    // En avslutande stjärna matchar resten - men bara
                    // om resten inte innehåller en separator.
                    return !misslyckad && !namn.contains(&'/');
                }

                // Sök efter en träff längre in i namnet. Stjärnan
                // får inte hoppa över en separator.
                let mut index = 0;

                while index < namn.len() && namn[index] != '/' {
                    if matchar_chunk(mönster, &namn[index + 1..]) {
                        return true;
                    }

                    index += 1;
                }

                return !misslyckad;
            }
            tecken => {
                if !misslyckad {
                    if namn.is_empty() || namn[0] != tecken {
                        misslyckad = true;
                    } else {
                        namn = &namn[1..];
                    }
                }

                mönster = &mönster[1..];
            }
        }
    }

    !misslyckad && namn.is_empty()
}

/// Kontrollerar om en sökväg finns i en lista.
pub fn finns(sökväg: &Path, lista: &[PathBuf]) -> bool {
    lista.iter().any(|post| post == sökväg)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn stjarnan_matchar_en_niva() {
        assert!(matchar_segment("*", "fil.go"));
        assert!(!matchar_segment("*", "mapp/fil.go"));
    }

    #[test]
    fn fragetecken_matchar_ett_tecken() {
        assert!(matchar_segment("?.go", "a.go"));
        assert!(!matchar_segment("?.go", "ab.go"));
    }

    #[test]
    fn teckenklass_matchar() {
        assert!(matchar_segment("[ab]*", "abc"));
        assert!(!matchar_segment("[ab]*", "cde"));
        assert!(matchar_segment("[a-c].go", "b.go"));
    }

    #[test]
    fn monster_utan_meta_finns_inte() {
        let träffar = glob("/tmp/plan-agent-glob-saknas/*.go").unwrap();

        assert!(träffar.is_empty());
    }
}
