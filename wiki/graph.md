<!-- GENERERAD av `wikictl graph`. Ändra inte manuellt. -->
# Länkgraph

Kunskapsgrafen: varje nod är en sida, varje kant är en länk mellan dem.

```mermaid
graph LR
  n_concepts_authentik["Authentik"]
  n_concepts_cert_manager["Cert-Manager"]
  n_concepts_cloudflare_tunnel["Cloudflare Tunnel (cloudflared)"]
  n_concepts_cloudnative_pg["CloudNativePG (CNPG)"]
  n_concepts_envoy_gateway["Envoy Gateway"]
  n_concepts_flux_operator["FluxOperator"]
  n_concepts_gitops["GitOps"]
  n_concepts_http_routing_flow["HTTP Routing Flow"]
  n_concepts_kube_prometheus["Kube Prometheus Stack"]
  n_concepts_kubelet_csr_approver["Kubelet CSR Approver"]
  n_concepts_kyverno_cel["Kyverno CEL Policies"]
  n_concepts_kyverno_policies["Kyverno Policies"]
  n_concepts_longhorn["Longhorn"]
  n_concepts_metallb["MetalLB"]
  n_concepts_neon_cli_var_neon_och_neon_ts_maste_ligga["Neon CLI: var .neon och neon.ts måste ligga"]
  n_concepts_oidc["OIDC"]
  n_concepts_poolad_eller_direkt_anslutning_mot_neon["Poolad eller direkt anslutning mot Neon"]
  n_concepts_sops["SOPS"]
  n_concepts_zitadel["Zitadel"]
  n_entities_flux_gitops["Flux GitOps"]
  n_entities_hdmi_kvm_switch_16_port_labbet["HDMI KVM-switch 16 port (labbet)"]
  n_entities_kubernetes["Kubernetes"]
  n_entities_pxe_server["PXE-server"]
  n_entities_talos_linux["Talos Linux"]
  n_notes_1password_mcp_environments_pa_arch_omarchy_kraver_setgid_pa_1password_mcp["1Password MCP (Environments) på Arch/Omarchy kräver setgid på 1password-mcp"]
  n_notes_accuracy_fel_omstart["Accuracy-felet vid omstart"]
  n_notes_admin_inloggning_via_losenord_i_stallet_for_token_i_ui_t["Admin-inloggning via lösenord i stället för token i UI:t"]
  n_notes_agent_api_crashade_i_tilt_saknad_database_url_mot_cnpg_databasen_i_agents_dev["Agent-api crashade i Tilt: saknad DATABASE_URL mot CNPG-databasen i agents-dev"]
  n_notes_agent_granssnittet_nuxt_saknades_helt_i_tilt_routes_forsvann_i_refaktoreringen_677b876["Agent-gränssnittet (Nuxt) saknades helt i Tilt - routes försvann i refaktoreringen 677b876"]
  n_notes_agentmal_flaggan_blev_levande_mvp["Agentmål: flaggan blev levande (MVP)"]
  n_notes_agentmarke_pa_mal_tangenten_a_ar_reserverad["Agentmärke på mål - tangenten a är reserverad"]
  n_notes_agentprofiler_som_teknisk_policy_utanfor_prompten["Agentprofiler som teknisk policy utanför prompten"]
  n_notes_arbetsstationen_ar_omarchy_arch_apt_get_finns_inte_anvand_pacman_yay["Arbetsstationen ar Omarchy/Arch - apt-get finns inte, anvand pacman/yay"]
  n_notes_arkitektur["Arkitektur"]
  n_notes_artikeln_om_agentic_software_factory_vad_som_ar_relevant_for_agentsystemet["Artikeln om agentic software factory - vad som är relevant för agentsystemet"]
  n_notes_aterkommande_mal_samma_malobjekt_med_upprepningsregel["Återkommande mål: samma målobjekt med upprepningsregel"]
  n_notes_aterkommande_serier_doljs_inte_langre_beroende_pa_zoomniva["Återkommande serier döljs inte längre beroende på zoomnivå"]
  n_notes_atta_feature_commits_pushade_till_det_arkiverade_mailmaster_repot["Åtta feature-commits pushade till det arkiverade mailmaster-repot"]
  n_notes_authflytten_i_plan_sessionen_ligger_nu_i_go_api_t_inte_i_nuxt["Authflytten i plan: sessionen ligger nu i Go API:t, inte i Nuxt"]
  n_notes_automatisk_malikon_behovde_sokord_och_svenska_alias["Automatisk målikon behövde sökord och svenska alias"]
  n_notes_autonoma_agentruns_som_kubernetes_jobs_lardomar_fran_e2e_flodet["Autonoma AgentRuns som Kubernetes Jobs: lärdomar från E2E-flödet"]
  n_notes_barnmarke_pa_mal_antal_undermal_visas_overallt["Barnmärke på mål - antal undermål visas överallt"]
  n_notes_barnordning_foljer_beroenden_i_malsidan["Barnordning följer beroenden i målsidan"]
  n_notes_bash_fallgropar_i_pxe_flash_tab_som_ifs_whitespace_och_read_med_for_fa_variabler["bash-fallgropar i pxe-flash: tab som IFS-whitespace och read med for få variabler"]
  n_notes_beroendevaljare_med_tangenten_w["Beroendeväljare med tangenten w"]
  n_notes_beroendevaljaren_kan_ta_bort_befintliga_beroenden["Beroendeväljaren kan ta bort befintliga beroenden"]
  n_notes_beslut_pi_durable_satts_pa_hyllan_for_tillfallet["Beslut: Pi Durable sätts på hyllan för tillfället"]
  n_notes_bigram_algoritmen["Bigram-algoritmen"]
  n_notes_boka_enskild_kommunbokning_i_matchi_fran_bookings["Boka enskild kommunbokning i MATCHi från /bookings"]
  n_notes_brevladelistan_maste_hamtas_innan_granskningslaget_startar["Brevlådelistan måste hämtas innan granskningsläget startar"]
  n_notes_cachade_imap_sessioner_kraver_hantering_av_doda_anslutningar["Cachade IMAP-sessioner kräver hantering av döda anslutningar"]
  n_notes_chattinterface_byggt_i_plan_slack_borttaget["Chattinterface byggt i Plan - Slack borttaget"]
  n_notes_ci_saknade_inloggning_mot_ghcr_alla_image_pushar_misslyckades["CI saknade inloggning mot GHCR - alla image-pushar misslyckades"]
  n_notes_cluster1_stale_kubeconfig_ca_och_noll_workers_diagnos_2026_11_11["cluster1: stale kubeconfig-CA och noll workers - diagnos 2026-11-11"]
  n_notes_controlplane_recovery_via_pxe_ar_inte_konfigurerad["Controlplane-recovery via PXE är inte konfigurerad"]
  n_notes_cookies_json_var_versionsstyrd_i_offentligt_neuro_repo["cookies.json var versionsstyrd i offentligt neuro-repo"]
  n_notes_detaljvy_for_kandidat_i_pappa_gemensamma_vagar_mellan_kandidater["Detaljvy för kandidat i pappa - gemensamma vägar mellan kandidater"]
  n_notes_dev_inloggning_i_plan_session_token_db_rattigheter_och_oauth_url_er["Dev-inloggning i plan: session-token, DB-rättigheter och OAuth-URL:er"]
  n_notes_dnsmasq_dhcp_boot_tar_bara_en_tagg_anvand_tag_if_for_villkor["dnsmasq --dhcp-boot tar bara EN tagg - använd tag-if för villkor"]
  n_notes_dold_skrivbordsmeny_med_leader_kommando["Dold skrivbordsmeny med leader-kommando"]
  n_notes_dolda_ikoner_i_barnmalsrader_pa_malsidan_visa_vid_hover["Dolda ikoner i målrader - visa bara det målet använder"]
  n_notes_en_talos_efi_partition_pa_kortet_blockerar_pxe_booten_bootflow_loaded_fore_natverket["En Talos-EFI-partition pa kortet blockerar PXE-booten (bootflow loaded fore natverket)"]
  n_notes_en_uttunnad_forekomst_maste_vara_synlig_aven_om_den_ligger_utanfor_fonstret["En uttunnad förekomst måste vara synlig även om den ligger utanför fönstret"]
  n_notes_enter_pa_goals_skapar_nytt_grundmal_istallet_for_att_oppna_malet["Enter på /goals skapar nytt grundmål istället för att öppna målet"]
  n_notes_enter_skapar_alltid_nytt_mal_piltangenterna_markerar_befintligt_mal_i_sokfalt["Enter skapar alltid nytt mål - piltangenterna markerar befintligt mål i sökfält"]
  n_notes_erfarenhetsfilter_i_pappa_it_experience_years_med_sparbart_grund["Erfarenhetsfilter i pappa: it_experience_years med spårbart grund"]
  n_notes_escape_i_datum_popup_maste_hanteras_fore_leader_listenern["Escape i datum-popup måste hanteras före leader-listenern"]
  n_notes_eviga_mal_i_plan_kan_inte_slutforas["Eviga mål i plan - kan inte slutföras"]
  n_notes_expanderade_eml_laggs_in_som_riktiga_mejl_i_mappen_append["Expanderade .eml läggs in som riktiga mejl i mappen (APPEND)"]
  n_notes_feedback_loop_ai_uppdaterar_mailreglerna_och_analyserar_om_mejlet["Feedback-loop: AI uppdaterar mailreglerna och analyserar om mejlet"]
  n_notes_figma_lokala_fonter_pa_linux_figma_agent_linux["Figma lokala fonter på Linux: figma-agent-linux"]
  n_notes_flera_kalendrar_med_egna_farger_i_plan["Flera kalendrar med egna färger i Plan"]
  n_notes_flerval_parkera_flera_kandidater_samtidigt_i_personer["Flerval: parkera flera kandidater samtidigt i /personer"]
  n_notes_foraldrar_syns_pa_mal_i_now["Föräldrar syns på mål i /now"]
  n_notes_globala_agentinstruktioner_deployas_via_dotfiles["Globala agentinstruktioner deployas via dotfiles"]
  n_notes_go_imap_uidstore_istallet_for_store_med_uid_mangd["go-imap: UidStore istället för Store med UID-mängd"]
  n_notes_go_integrationstest_tomde_kalenderhandelserna_i_neon_igen["Go-integrationstest tömde kalenderhändelserna i Neon igen"]
  n_notes_goals_listar_bara_grundmal_scroll_vid_j_k["/goals listar bara grundmål + scroll vid j/k"]
  n_notes_gotcha_pxe_forst_gor_noden_till_en_pxe_nod_maintenance_mode["Gotcha: 'PXE först' gör noden till en PXE-nod (maintenance mode)"]
  n_notes_gron_markering_av_tider_som_ar_bokade_i_matchi_skrivlasande_jamforelse["Grön markering av tider som är bokade i MATCHi (skrivläsande jämförelse)"]
  n_notes_ha_automationer_behover_en_startv_g_for_att_aterstalla_laget_och_tre_syntaxfallor["HA: automationer behöver en startvág för att återställa läget (och tre syntaxfällor)"]
  n_notes_handelsen_har_ingen_egen_farg["Händelsen har ingen egen färg"]
  n_notes_hierarki_i_gor_nu_listan_now["Hierarki i Gör nu-listan (/now)"]
  n_notes_hjalpmodalen_for_kortkommandon_fick_sokfilter["Hjälpmodalen för kortkommandon fick sökfilter"]
  n_notes_human_gate_som_teknisk_grans_scope_pa_api_nyckeln["Human gate som teknisk gräns: scope på API-nyckeln"]
  n_notes_ikonsokning_far_timeout_och_checkbox_fallback_vid_nya_mal["Ikonsökning får timeout och checkbox-fallback vid nya mål"]
  n_notes_ikonvaljaren_i_plan_api_icons_routen_var_borttagen_och_ikonprovider_none_gjorde_att_inga_ikoner_renderades["Ikonväljaren i plan: /api-icons-routen var borttagen och ikonprovider none gjorde att inga ikoner renderades"]
  n_notes_imap_inloggning_for_michael_brundin_soulmate_se_kraver_admin_consent["IMAP-inloggning för michael.brundin@soulmate.se kräver admin consent"]
  n_notes_imap_sessionen_i_mailagenten_ar_inte_tradsaker_parallella_listanrop_forstor_anslutningen["IMAP-sessionen i mailagenten är inte trådsäker - parallella listanrop förstör anslutningen"]
  n_notes_inaktiva_ikon_rader_dolda_helt_ingen_graad_variant_vid_hover["Inaktiva ikon-rader dolda helt - ingen gråad variant vid hover"]
  n_notes_infrastructure_agents_md_uppdaterad_tre_fallor_for_agenter["infrastructure: AGENTS.md uppdaterad + tre fällor för agenter"]
  n_notes_interaktiv_malchatt_behover_sakra_plan_verktyg["Interaktiv målchatt behöver säkra Plan-verktyg"]
  n_notes_ipv6_ingress_till_envoy_designlage_2026_11_15["'IPv6-ingress till Envoy: designläge 2026-11-15 (klustret är nere)'"]
  n_notes_it_arbetsplatser_i_sundsvall_3_469_enligt_scb_yrkesregistret["IT-arbetsplatser i Sundsvall: 3 469 enligt SCB Yrkesregistret"]
  n_notes_jobb_ska_avslutas_nar_korningen_ar_klar_och_en_delad_proxy_rensar_typescript_felen["Jobb ska avslutas när körningen är klar, och en delad proxy rensar TypeScript-felen"]
  n_notes_kalender_som_permanent_hogerspalt_pa_desktop["Kalender som permanent högerspalt på desktop"]
  n_notes_kalenderhandelser_aterstallda_fran_neon_wal_historik["Kalenderhändelser återställda från Neon WAL-historik"]
  n_notes_kalenderhandelser_kan_ha_egna_farger["Kalenderhändelser kan ha egna färger"]
  n_notes_kalenderhandelser_kan_tas_bort_och_dupliceras["Kalenderhändelser kan tas bort och dupliceras"]
  n_notes_kalendern_bara_etikettbandet_reserverar_plats_i_kolumnen["Kalendern: bara etikettbandet reserverar plats i kolumnen"]
  n_notes_kalendern_begransar_aterkommande_forekomster_i_langtidvyer["Kalendern begränsar återkommande förekomster i långtidvyer"]
  n_notes_kalendern_begransas_till_tre_manader["Kalendern begränsas till tre månader"]
  n_notes_kalendern_en_kolumn_dolda_serier_och_en_installningssida["Kalendern: en kolumn, dolda serier och en inställningssida"]
  n_notes_kalendern_hamtar_alla_handelser_en_gang_och_expanderar_serier_lokalt["Kalendern hämtar alla händelser en gång och expanderar serier lokalt"]
  n_notes_kalendern_sammanfattningen_av_serier_togs_bort["Kalendern: sammanfattningen av serier togs bort"]
  n_notes_kalendern_urval_istallet_for_dold_serie["Kalendern: urval istället för dold serie"]
  n_notes_kalendern_urvalet_togs_bort_serien_doljs_som_helhet["Kalendern: urvalet togs bort - serien döljs som helhet"]
  n_notes_kalendern_visar_handelser_som_rader_aldrig_som_prickar["Kalendern visar händelser som rader, aldrig som prickar"]
  n_notes_kalenderns_detaljnivaer_fem_nivaer_och_lane_packning["Kalenderns detaljnivåer: fem nivåer och lane-packning"]
  n_notes_kalenderns_handelser_lag_over_modaler_och_escape_var_dovt["Kalenderns händelser låg över modaler och Escape var dövt"]
  n_notes_kalenderns_nivavaxling_animeras_crossfade_pop_in_och_hysteresis["Kalenderns nivåväxling animeras: crossfade, pop-in och hysteresis"]
  n_notes_kalenderns_scroll_axlar_bytta_lodrat_zoomar_vagratt_panorerar["Kalenderns scroll-axlar bytta: lodrät zoomar, vågrätt panorerar"]
  n_notes_kalenderns_tangentbordskontroller_ska_inte_stora_malnavigering["Kalenderns tangentbordskontroller ska inte störa målnavigering"]
  n_notes_kalenderns_tidsaxelrader_animeras_vid_zoom["Kalenderns tidsaxelrader animeras vid zoom"]
  n_notes_kalenderns_zoom_styrs_av_vertikal_scrollning["Kalenderns zoom styrs av vertikal scrollning"]
  n_notes_kalenderrutorna_har_ingen_maxhojd_langre["Kalenderrutorna har ingen maxhöjd längre"]
  n_notes_kameo_beloppet_kommer_fran_transaktionerna_inte_fran_mejlet["Kameo-beloppet kommer fran transaktionerna, inte fran mejlet"]
  n_notes_kameo_ranteutbetalning_redan_utbetald_samlingsuttag_pa_47_29_kr["Kameo ranteutbetalning redan utbetald - samlingsuttag pa 47,29 kr"]
  n_notes_kameo_sidan_flyttad_till_toppniva_och_delad_i_flikar["Kameo-sidan flyttad till toppnivå och delad i flikar"]
  n_notes_kameo_uppgifter_synkas_till_fil_sa_docker_slipper_op_cli["Kameo-uppgifter synkas till fil så Docker slipper op-cli"]
  n_notes_kameo_utbetalningar_nuxt_sida_som_laser_lan_fran_kameo["Kameo utbetalningar: Nuxt-sida som läser lån från Kameo"]
  n_notes_kanata_switch_session_tmuxinator_fonster_oppnades_inte["Kanata switch-session: tmuxinator-fönster öppnades inte"]
  n_notes_kandidatlistan_heter_nu_personer["Kandidatlistan heter nu /personer"]
  n_notes_klustret_anvander_sops_via_flux_inte_vault_eso["Klustret anvander SOPS via Flux - inte Vault/ESO"]
  n_notes_klustret_har_byggts_om_bara_kube_system_flux_och_appar_saknas_2026_11_23["Klustret har byggts om - bara kube-system, Flux och appar saknas (2026-11-23)"]
  n_notes_koda_korningar_ar_uttryckliga_begaran_och_en_avslutad_korning_har_alltid_run_completed["Köda körningar är uttryckliga begäran - och en avslutad körning har alltid run_completed"]
  n_notes_konsulter_i_sundsvall_antalet_gar_inte_att_mata_det_som_gar_att_saga["'Konsulter i Sundsvall: antalet går inte att mäta - det som går att säga'"]
  n_notes_konsulter_vs_anstallda_i_sundsvall_halva_it_styrkan_sitter_i_konsultnara_bolag["'[DEPRECATED] 50/50-uppdelning mellan konsulter och anställda - FEL, se den nya noten'"]
  n_notes_konventioner["Konventioner i repot"]
  n_notes_korrigering_rock4c_plus_bin_ar_ett_microsd_emmc_image_inte_en_spi_image["Korrigering: rock4c-plus.bin är ett microSD/eMMC-image, inte en SPI-image"]
  n_notes_kortkommando_hjalpen_i_helpmodal_maste_folja_tangentandringar["Kortkommando-hjälpen i HelpModal måste följa tangentändringar"]
  n_notes_kostnad_och_budget_for_agentkorningar["Kostnad och budget för agentkörningar"]
  n_notes_krypterade_secrets_i_klustret_ogiltiga_inloggningar_och_keycloak_brute_force["Krypterade secrets i klustret: ogiltiga inloggningar och Keycloak-brute force"]
  n_notes_kunskaps_wiki["Kunskaps-wikin"]
  n_notes_kureringspasset_hittade_foraldrad_kunskap_i_wikin["Kureringspasset hittade foraldrad kunskap i wikin"]
  n_notes_kvm_portmappning_worker_n_port_n_den_deployade_configen_pa_pi_n_var_fel["KVM-portmappning: worker-N = port N - den deployade configen pa Pi:n var fel"]
  n_notes_kvm_styrning_fran_pxe_pi_n_16_port_hdmi_kvm["KVM-styrning från PXE-Pi:n (16-port HDMI KVM)"]
  n_notes_leader_d_oppnar_gemensam_datum_popup_for_mal["Leader D öppnar gemensam datum-popup för mål"]
  n_notes_leader_key_gjordes_global_mellanslag_sekvenser_pa_alla_sidor["Leader-key gjordes global (mellanslag + sekvenser på alla sidor)"]
  n_notes_leader_laget_kunde_inte_oppnas_efter_musklick_pa_en_knapp["Leader-läget kunde inte öppnas efter musklick på en knapp"]
  n_notes_learn_calibre_synk_och_automatisk_import_via_systemd["Learn: Calibre-synk och automatisk import via systemd"]
  n_notes_learn_designsystem_a3_folio_i_nuxt["Learn: designsystem A3 Folio i Nuxt"]
  n_notes_learn_korsanalys_av_varje_bok_mot_den_samlade_kunskapsbasen["Learn: korsanalys av varje bok mot den samlade kunskapsbasen"]
  n_notes_learn_kunskapsbasen_kan_lagras_i_postgresql["Learn: kunskapsbasen kan lagras i PostgreSQL"]
  n_notes_learn_kunskapssystem_implementerat_i_src_python["Learn: kunskapssystem implementerat i src/python"]
  n_notes_learn_kunskapssystemet_portat_fran_python_till_go["Learn: kunskapssystemet portat från Python till Go"]
  n_notes_learn_rest_api_och_nuxt_spa_mot_samma_kunskapsbas["Learn: REST-API och Nuxt-SPA mot samma kunskapsbas"]
  n_notes_lightcraft_photocraft_installerade_pa_omarchy_workstation_utan_sudo["LightCraft + PhotoCraft installerade på Omarchy-workstation (utan sudo)"]
  n_notes_linkedin_lankar_for_217_av_230_personer_13_kvarvarande_oroliga["LinkedIn-länkar för 217 av 230 personer - 13 kvarvarande oroliga"]
  n_notes_linkedin_relationer_i_pappa_422_svaga_samband_inlagda_och_linkedin_bot_sp_u00e4rr_kontot["LinkedIn-relationer i pappa - 422 svaga samband inlagda, och LinkedIn bot-sp\u00e4rr kontot"]
  n_notes_lokal_expansion_kan_verifieras_mot_serverns_expansion["Lokal expansion kan verifieras mot serverns expansion"]
  n_notes_losning_1password_backed_env_variabel_for_mcp_tokens["Lösning: 1Password-backed env-variabel för MCP-tokens"]
  n_notes_lunchdeploy_triggas_fran_cronjob_pa_pxe["Lunchdeploy triggas från cronjob på pxe"]
  n_notes_lunchmenyn_delar_en_modell_mellan_byggskript_och_webblasare["Lunchmenyn delar en modell mellan byggskript och webblasare"]
  n_notes_mailagent_bygget_brot_pa_ofardig_kameo_integration["Mailagent-bygget bröt på ofärdig Kameo-integration"]
  n_notes_mailagenten_expanderar_bifogade_eml_till_egna_mejl["Mailagenten expanderar bifogade .eml till egna mejl"]
  n_notes_mailagenten_flyttad_fran_mailmaster_repot_in_i_pappa["Mailagenten flyttad från mailmaster-repot in i pappa"]
  n_notes_mailagentens_kontodropdown_renderades_aldrig_nuxt_prefixade_komponenter["Mailagentens kontodropdown renderades aldrig - Nuxt-prefixade komponenter"]
  n_notes_mailagentens_kontofel_gav_alltid_http_500_och_listan_foljde_inte_kontot["Mailagentens kontofel gav alltid HTTP 500 och listan följde inte kontot"]
  n_notes_mailagentens_mcp_verktygen_saknade_konto_och_hade_fel_schema["Mailagentens MCP-verktygen saknade konto och hade fel schema"]
  n_notes_mailgranskningen_visar_ett_mejl_i_taget["Mailgranskningen visar ett mejl i taget"]
  n_notes_mailmaster_api_v1_gav_failed_to_fetch_go_servern_panickade_och_saknade_cors["MailMaster: /api/v1 gav 'Failed to fetch' - Go-servern panickade och saknade CORS"]
  n_notes_mailmaster_gammal_container_build_korde_process_och_arkiverade_mejl["MailMaster: gammal container-build körde /process och arkiverade mejl"]
  n_notes_mailmaster_granskning_av_ai_atgarder_fore_utforande["MailMaster: granskning av AI-åtgärder före utförande"]
  n_notes_mailmaster_kontodropdown_outlook_inloggning_via_device_code["MailMaster: kontodropdown + Outlook-inloggning via device code"]
  n_notes_mailmaster_svarmejl_fick_fel_brodtext_tre_buggar_i_svarttextskedjan["MailMaster: svarmejl fick fel brödtext - tre buggar i svarttextskedjan"]
  n_notes_mailvyn_mail_omstegad_till_tailwind_tema["Mailvyn (/mail) omstegad till Tailwind-tema"]
  n_notes_mal_kan_ha_due_datum_deadline["Mål kan ha due-datum (deadline)"]
  n_notes_mal_kan_ha_en_lang_beskrivning_redigerad_med_tangenten_t["Mål kan ha en lång beskrivning, redigerad med tangenten t"]
  n_notes_mal_kan_ha_startdatum_syns_i_priority_forst_nar_datumet_passerat["Mål kan ha startdatum - syns i /priority först när datumet passerat"]
  n_notes_mal_kan_schemalaggas_och_kalenderhandelser_flyttas["Mål kan schemaläggas och kalenderhändelser flyttas"]
  n_notes_malfaserna_ar_gemensamma_for_alla_mal_i_plan["Målfaserna är gemensamma för alla mål i Plan"]
  n_notes_malgraph_navigering_mellan_malsidor_gar_fran_sekunder_till_millisekunder["Målgraph: navigering mellan målsidor går från sekunder till millisekunder"]
  n_notes_malikoner_anvander_enkel_oversattning_fore_iconify["Målikoner använder enkel översättning före Iconify"]
  n_notes_manuell_sortering_av_mal_pa_priority["Manuell sortering av mål på priority"]
  n_notes_marknadslage_panel_hogst_upp_pa_personer_i_pappa["Marknadsläge-panel högst upp på /personer i pappa"]
  n_notes_masterdatabas_i_postgres_schema_import_och_drift_for_it_rekryteringen["Masterdatabas i Postgres: schema, import och drift för IT-rekryteringen"]
  n_notes_matchi_bokning_slot_id_andras_nar_bokningen_skapas["MATCHi-bokning: slot-id ändras när bokningen skapas"]
  n_notes_matchi_inloggningen_ar_flaky_klienten_forsoker_om_och_dev_klustret_behover_ombyggnad["MATCHi-inloggningen är flaky - klienten försöker om, och dev-klustret behöver ombyggnad"]
  n_notes_matchi_keycloak_session_code_maste_fangas_med_generell_teckenklass["MATCHi Keycloak: session_code måste fångas med generell teckenklass"]
  n_notes_md_viewer_tauri_installerades_men_avinstallerades_igen["MD Viewer (Tauri) installerades men avinstallerades igen"]
  n_notes_median_i_stallet_for_snitt_for_wpm["Median i stället för snitt för WPM"]
  n_notes_mejlbrodtext_renderas_som_markdown_i_granskningsvyn["Mejlbrödtext renderas som markdown i granskningsvyn"]
  n_notes_mellanslag_blockerades_i_malbeskrivningen["Mellanslag blockerades i målbeskrivningen"]
  n_notes_mellanslag_i_radxa_rock_4c_saboterar_byggen_losning_bind_mount["Mellanslag i 'Radxa Rock 4C+' saboterar byggen - losning: bind-mount"]
  n_notes_michael_brundin_blir_en_personnod_i_relationsgrafen["Michael Brundin blir en personnod i relationsgrafen"]
  n_notes_min_kontaktlista_jag_som_nod_i_grafen_pappa["Min kontaktlista - jag som nod i grafen (pappa)"]
  n_notes_moderposten_arkiveras_nar_eml_bilagorna_har_lagts_in["Moderposten arkiveras när .eml-bilagorna har lagts in"]
  n_notes_neon_api_token_hamnade_i_dotfiles_git_historiken["Neon API-token hamnade i dotfiles-git-historiken"]
  n_notes_neuroquant_ar_ombyggt_trpc_api_och_e_post_verifiering["NeuroQuant är ombyggt - tRPC-API och e-post-verifiering"]
  n_notes_neuroquant_kontot_last_av_for_manga_inloggningsforsok["NeuroQuant-kontot låst av för många inloggningsförsök"]
  n_notes_neuroquant_vikterna_raknas_om_dagligen_mejl_triggas_pa_publicering["NeuroQuant-vikterna räknas om dagligen - mejl triggas på publicering"]
  n_notes_nu_marke_pa_mal_leader_n_styr_gor_nu_listan["Nu-märke på mål (leader+N) styr Gör nu-listan"]
  n_notes_nuxt_4_webbapp_for_pappa_laser_personerna_ur_neon["Nuxt 4-webbapp för pappa läser personerna ur Neon"]
  n_notes_nuxt_cli_via_bun_run_dev_gar_sonder_i_stk["Nuxt-CLI via bun run dev går sönder i stk"]
  n_notes_nuxt_sidomeny_kandidater_som_startsida_i_pappa["Nuxt-sidomeny + /kandidater som startsida i pappa"]
  n_notes_ny_rs232_hardvara_ak_co_02_db9_breakout_och_kopplingen_till_kvm_ns_3_pin["Ny RS232-hårdvara: AK-CO-02 + DB9-breakout, och kopplingen till KVM:ns 3-pin"]
  n_notes_ny_terminal_i_tmux_fragade_efter_datorns_losenord["Ny terminal i tmux frågade efter datorns lösenord"]
  n_notes_nya_mal_far_ikon_fran_titeln["Nya mål får ikon från titeln"]
  n_notes_omarchy_idle_lasning_stangs_av_via_shell_json_inte_hypridle_conf["Omarchy: idle-låsning stängs av via shell.json (inte hypridle.conf)"]
  n_notes_one_com_lagrar_mappar_i_nfd_normalisera_i_getfoldername["one.com lagrar mappar i NFD - normalisera i getFolderName"]
  n_notes_p_tangenten_konfliktade_med_viktlaget_pa_priority["p-tangenten konfliktade med viktläget på /priority"]
  n_notes_pa_rock_plattorna_finns_ingen_bootloader_i_hardvaran["På Rock-plattorna finns ingen bootloader i hårdvaran"]
  n_notes_pappa_admin_token_ett_stalle_for_token_inte_handpallagning_per_miljo["pappa: admin-token - ett stalle for token, inte handpallagning per miljo"]
  n_notes_pappa_anstallningshistorik_i_cv_format_fix_av_kompetensordboken["pappa: anställningshistorik i CV-format + fix av kompetensordboken"]
  n_notes_pappa_dubblettknapp_mansklig_granskning_och_bestandig_sammanslagning["pappa: dubblettknapp - mänsklig granskning, delvis matchning och beständig sammanslagning"]
  n_notes_pappa_hemlighet_i_build_context_env_i_docker_imagen["pappa: hemlighet i build-context - .env i Docker-imagen"]
  n_notes_pappa_kor_nu_pa_neon_neon_ts_env_local_och_poolad_direkt_wiring["pappa kör nu på Neon: neon.ts, .env.local och poolad/direkt-wiring"]
  n_notes_pappa_regelendpointet_laser_och_sparar_riktig_rules_md_var_stub["pappa: regelendpointet läser och sparar riktig rules.md (var stub)"]
  n_notes_pappa_uppdragsmatchning_delpoang_forklarande_ranking["pappa: uppdragsmatchning - delpoang forklarande ranking"]
  n_notes_pappa_vyn_uppdrag_kravformular_och_rankad_kandidatlista["pappa: vyn Uppdrag - kravformulär och rankad kandidatlista"]
  n_notes_pappa_yaml_frolistan_kan_aterskapas_ur_masterdatabasen["pappa: YAML-frölistan kan återskapas ur masterdatabasen"]
  n_notes_pappas_dev_miljo_kor_mot_neon_i_stallet_for_in_kluster_postgres["pappas dev-miljö kör mot Neon i stället för in-kluster-Postgres"]
  n_notes_pappas_grundare_ar_michael_brundin_inte_git_anvandaren_simon_brundin["Pappas grundare är Michael Brundin - inte git-användaren Simon Brundin"]
  n_notes_parallella_agent_sessioner_i_samma_repo_och_databas["Parallella agent-sessioner i samma repo och databas"]
  n_notes_personer_kan_laggas_in_manuellt_i_pappa["Personer kan läggas in manuellt i pappa"]
  n_notes_photocraft_agentstyrning_via_control_protokoll_control_7878["PhotoCraft agentstyrning via control-protokoll (--control 7878)"]
  n_notes_photocraft_kamerans_look_varning_vid_raw_oppning_och_hur_den_hanteras["PhotoCraft: kamerans look-varning vid RAW-öppning och hur den hanteras"]
  n_notes_pi_desktop_notis_nar_agent_ar_klar["Pi desktop-notis när agent är klar"]
  n_notes_pi_durable_1_4_0_api_avvikelser_fran_blogginlagget_och_tva_gotchas["Pi Durable 1.4.0: API-avvikelser från blogginlägget och två gotchas"]
  n_notes_pi_durable_demo_paketet_earendil_works_pi_durable_fungerar["Pi Durable-demo (paketet @earendil-works/pi-durable fungerar)"]
  n_notes_pi_durable_fungerar_med_minimax_coding_plan_anthropic_messages_api["Pi Durable fungerar med MiniMax coding plan (anthropic-messages-API)"]
  n_notes_pi_tog_bort_pi_rate_limit_och_pi_pattern_retry_fixade_omarchy_temasy["Pi: tog bort pi-rate-limit och pi-pattern-retry, fixade omarchy temasy"]
  n_notes_picokvm_luckfox_startar_inte_i_labbet_lite_saknar_emmc_ingen_poe["PicoKVM (Luckfox) startar inte i labbet: Lite saknar eMMC + ingen PoE"]
  n_notes_plan_api_anvander_scalar_for_openapi["Plan API använder Scalar för OpenAPI"]
  n_notes_plan_appen_morkt_designsystem_med_semantiska_farg_token["Plan-appen: mörkt designsystem med semantiska färg-token"]
  n_notes_plan_cachad_spa_skal_gjorde_att_webblasaren_korde_gammalt_js_bundle_med_ikon_provider_none["Plan: cachad SPA-skal gjorde att webbläsaren körde gammalt JS-bundle med ikon-provider none"]
  n_notes_plan_dev_kor_mot_neon_istallet_for_lokal_postgres["Plan dev kör mot Neon istället för lokal Postgres"]
  n_notes_plan_repots_ci_och_prod_pipeline_var_trasig_fixade_sokvagar_och_bygg["Plan-repots CI- och prod-pipeline var trasig - fixade sökvägar och bygg"]
  n_notes_platser_pa_mal_och_kalenderhandelser_openspec_change_add_places["Platser på mål och kalenderhändelser (OpenSpec-change add-places)"]
  n_notes_popcorn_time_installerad_som_appimage_i_local["Popcorn Time installerad som AppImage i ~/.local"]
  n_notes_popcorn_time_remote_movies_api_failed_to_respond_orsak_och_fix["Popcorn Time: 'remote movies API failed to respond' - orsak och fix"]
  n_notes_popcorn_time_svenskt_ljud_finns_inte_i_katalogen_anvand_torrentsamling_torrent_collection["Popcorn Time: svenskt ljud finns INTE i katalogen - använd Torrentsamling (Torrent Collection)"]
  n_notes_popup_editor_i_klickbar_rad_klicken_bubblar_och_navigerar_bort["Popup-editor i klickbar rad: klicken bubblar och navigerar bort"]
  n_notes_prioriteringsordningen_pa_now_sparades_inte_bulk_update_och_spar_ko["Prioriteringsordningen på /now sparades inte - bulk-update och spar-kö"]
  n_notes_priority_vyn_byter_namn_till_gor_nu["Priority-vyn byter namn till Gör nu"]
  n_notes_priorityfilter_ska_behandla_saknat_started_som_ej_paborjat["Priorityfilter ska behandla saknat started som ej påbörjat"]
  n_notes_pxe_aterstallning_nodspecifika_matchbox_profiler_och_maintenance_mode_flode["PXE-återställning: nodspecifika matchbox-profiler och maintenance-mode-flöde"]
  n_notes_pxe_boot_script_dhcp_kraver_mkimage_uimage_inte_ren_hush_text["PXE: boot_script_dhcp kräver mkimage-uImage, inte ren hush-text"]
  n_notes_pxe_bootloop_initramfs_startar_om_for_att_boota_installerad_talos_pa_emmc["PXE-bootloop: initramfs startar om for att boota installerad Talos pa eMMC"]
  n_notes_pxe_flash_hanterar_divergerat_infrastructure_repo["PXE-flash hanterar divergerat infrastructure-repo"]
  n_notes_pxe_forst_bootloader_for_rock_4c_byggd_som_patchad_u_boot_image["PXE-först-bootloader för ROCK 4C+ byggd som patchad U-Boot-image"]
  n_notes_pxe_genererade_nodconfigs_var_stale_efter_talos_v1_14_uppgraderingen["PXE-genererade nodconfigs var stale efter Talos v1.14-uppgraderingen"]
  n_notes_pxe_kedjan_dhcp_boot_pekar_pa_textskript_matchbox_efi_ar_uefi_vagen["PXE-kedjan: dhcp-boot pekar på textskript, matchbox.efi är UEFI-vägen"]
  n_notes_pxe_kedjan_for_worker_1_verifierad_serverside_pxe_metoden_failar_snabbt_utan_pxelinux_cfg["PXE-kedjan for worker-1 verifierad serverside; pxe-metoden failar snabbt utan pxelinux.cfg"]
  n_notes_pxe_mainline_u_boots_bootefi_kraver_addr_size_for_image_och_initrd["PXE: mainline-U-Boots bootefi kräver addr:size for image och initrd"]
  n_notes_pxe_status_fungerar_i_produktionsklustret_tftp_svarar_inte_fran_pod_natet["PXE-status fungerar i produktionsklustret - TFTP svarar inte fran pod-natet"]
  n_notes_pxe_status_pa_dashboard_startsidan_tftp_gar_inte_att_kontrollera_fran_pod_natet["PXE-status på dashboard-startsidan: TFTP går inte att kontrollera från pod-nätet"]
  n_notes_pxe_varden_korde_foraldrad_matchbox_nodconfiger_saknades_404["PXE-värden körde föråldrad matchbox - nodconfiger saknades (404)"]
  n_notes_qbittorrent_ar_standardhanterare_for_torrent_och_magnet_lankar["qBittorrent är standardhanterare för .torrent och magnet-länkar"]
  n_notes_qbittorrent_installerad_pa_workstation_omarchy_arch["qBittorrent installerad på workstation (Omarchy/Arch)"]
  n_notes_query_parametrar_i_actor_smartbook_endpointen_url_avkodas_till_mellanslag["Query-parametrar i Actor Smartbook-endpointen: '+' URL-avkodas till mellanslag"]
  n_notes_radxa_bradena_behover_bootmedium_men_talos_och_bootloader_maste_ligga_pa_skilda_medium["Radxa-bradena behover bootmedium - men Talos och bootloader maste ligga pa skilda medium"]
  n_notes_rantebetalningsmejl_hamnar_i_inbox_deleted_messages_och_arkivet_matchades_aldrig["Rantebetalningsmejl hamnar i INBOX.Deleted Messages - och arkivet matchades aldrig"]
  n_notes_recovery_av_en_rock_worker_kraver_fysisk_rattning_esp_filen_pa_emmc_fangar_booten["Recovery av en ROCK-worker kräver fysisk rättning: ESP-filen på eMMC fångar booten"]
  n_notes_redigerare_och_agarskap_for_relationer_i_pappa["Redigerare och ägarskap för relationer i pappa"]
  n_notes_reglerna_i_pappa_lag_i_containerns_filysystem_sparade_regler_forsvann_vid_omstart["Reglerna i pappa låg i containerns filysystem - sparade regler försvann vid omstart"]
  n_notes_rekrytering_av_it_personal_i_sundsvall_marknadsbild_och_frolista["Rekrytering av IT-personal i Sundsvall: marknadsbild och frölista"]
  n_notes_rekryteringspipeline_board_per_person_sourcing_som_attribut["Rekryteringspipeline: board per person, sourcing som attribut"]
  n_notes_relationer_kan_nu_laggas_till_fran_oversikten_dialogen_har_par_lage["Relationer kan nu läggas till från översikten - dialogen har par-läge"]
  n_notes_relationer_mellan_kandidater_i_pappa_symmetriska_par_med_styrka_1_10["Relationer mellan kandidater i pappa - symmetriska par med styrka 1-10"]
  n_notes_relationsgrafen_anvander_fast_michael_id_utan_person_kind["Relationsgrafen använder fast Michael-id utan person_kind"]
  n_notes_relationsoversikt_i_pappa_var_ser_jag_mina_egna_relationer["Relationsöversikt i pappa - var ser jag mina egna relationer"]
  n_notes_repo_template_ci_gitops_flodet_fixat_bygger_3_images_ghcr_tagging_pr_overlays_per_pr["repo-template: CI/GitOps-flödet fixat (bygger 3 images, GHCR-tagging, PR-overlays per PR)"]
  n_notes_repo_template_ny_rust_api_src_rust_med_axum_utoipa_scalar["repo-template: ny Rust-API (src/rust) med Axum + utoipa + Scalar"]
  n_notes_repo_template_rust_api_kopplad_till_kubernetes_manifests_tilt["repo-template: Rust-API kopplad till Kubernetes (manifests + Tilt)"]
  n_notes_rescue_wipe_pa_rock_maste_tomma_hela_systemdisken_wipe_system_esp_filen_ar_det_som_fangar_booten["Rescue-wipe pa ROCK maste tomma hela systemdisken (wipe=system) - ESP-filen ar det som fangar booten"]
  n_notes_rk3399_s_bootrom_soker_spi_emmc_sd_usb_hardkodat["RK3399:s BootROM söker SPI → eMMC → SD → USB (hårdkodat)"]
  n_notes_rock_4c_4se_talos_kan_aldrig_installeras_till_disk_u_boot_saknar_efi_setvariablert["ROCK 4C+/4SE: Talos kan aldrig installeras till disk (U-Boot saknar EFI SetVariableRT)"]
  n_notes_rock_4c_4se_utan_batteribackup_rtc_star_pa_2017_tls_mot_factory_talos_dev_misslyckas["Rock 4C+/4SE utan batteribackup: RTC står på 2017, TLS mot factory.talos.dev misslyckas"]
  n_notes_rock_4c_och_4se_har_ingen_spi_flash_i_fabrik_u_boot_ligger_pa_microsd_emmc["ROCK 4C+ och 4SE har ingen SPI-flash i fabrik - U-Boot ligger på microSD/eMMC"]
  n_notes_rock_pi_4_rk3399_defconfig_bygger_en_rock_pi_4a_for_4c_ska_rock_4c_plus_rk3399_defconfig_anvandas["rock-pi-4-rk3399_defconfig bygger en ROCK Pi 4A - for 4C+ ska rock-4c-plus-rk3399_defconfig anvandas"]
  n_notes_root_dir_funktionen_stoppar_no_project_root_found_meddelandet["root_dir-funktionen stoppar 'No project root found'-meddelandet"]
  n_notes_rust_analyzer_standalone_lage_kraschar_pa_leetcode_nvim_filer["rust-analyzer standalone-läge kraschar på leetcode.nvim-filer"]
  n_notes_rust_analyzer_tysta_health_status_notisen_istallet_for_att_stanga_av_standalone["rust-analyzer: tysta health-status-notisen istället för att stänga av standalone"]
  n_notes_rust_learning_repot_lades_in_som_submodule_mot_egen_fork["rust: learning-repot lades in som submodule mot egen fork"]
  n_notes_rust_repo_learn_to_code_with_rust_som_submodule["rust-repo: learn-to-code-with-rust som submodule"]
  n_notes_rust_utvecklingsmiljo_pa_workstationen["Rust-utvecklingsmiljö på workstationen"]
  n_notes_sa_dampas_bada_rust_analyzer_popupen_no_project_root_found["Så dämpas båda rust-analyzer-popupen (No project root found)"]
  n_notes_sales_soulmate_se_appen_kan_inte_logga_in_oauth_clienten_saknas_i_soulmate_tenanten["sales@soulmate.se: appen kan inte logga in (OAuth-clienten saknas i soulmate-tenanten)"]
  n_notes_sammanslagning_av_agents_repot_in_i_plan_vad_som_flyttades_och_vad_som_medvetet_inte_gjorde_det["Sammanslagning av agents-repot in i plan: vad som flyttades och vad som medvetet inte gjorde det"]
  n_notes_serveragd_ko_for_stegvis_mailgranskning["Serverägd kö för stegvis mailgranskning"]
  n_notes_session_livscykel["Sessionens livscykel"]
  n_notes_sessionen_saknade_executor_varje_godkand_atgard_kraschade_med_502["Sessionen saknade executor - varje godkänd åtgärd kraschade med 502"]
  n_notes_simon_kvm_cli_kommando_for_att_byta_hdmi_kvm_port_fran_pxe_vard["simon kvm: CLI-kommando för att byta HDMI-KVM-port från PXE-värd"]
  n_notes_simon_pxe_flash_clean_code_omgang_2_fokuserade_funktioner["simon pxe flash: clean code-omgång 2 - fokuserade funktioner"]
  n_notes_simon_pxe_flash_flashar_pxe_forst_u_boot_till_sd_kort_i_pxe_servern["simon pxe flash - flashar PXE-först-U-Boot till SD-kort i PXE-servern"]
  n_notes_simonbrundin_gmail_com_inkorgen_kan_lasas_via_imap_begbot_losenordet["simonbrundin@gmail.com-inkorgen kan läsas via IMAP (begbot-lösenordet)"]
  n_notes_skogsduvan_elementet_styrs_av_minsta_temperaturen_fail_open["Skogsduvan: Elementet styrs av minsta temperaturen (fail-open)"]
  n_notes_skogsduvan_fyra_shelly_enheter_svarar_inte_pa_natet["Skogsduvan: fyra Shelly-enheter svarar inte på nätet"]
  n_notes_skogsduvan_larm_nar_sensorer_shelly_enheter_faller_ur_online["Skogsduvan: larm när sensorer/Shelly-enheter faller ur online"]
  n_notes_skogsduvan_ssh_via_duckdns_fungerar_port_22_forwardad_2026_11_02["'Skogsduvan: SSH via duckdns fungerar (port 22 forwardad 2026-11-02)'"]
  n_notes_skogsduvan_temperaturkaskad_i_tre_steg_styrd_av_minsta_temperaturen["Skogsduvan: temperaturkaskad i tre steg, styrd av minsta temperaturen"]
  n_notes_skogsduvans_ha_temperaturlogik_ombyggd_fyra_doda_sensorer["Skogsduvans HA: temperaturlogik ombyggd, fyra döda sensorer"]
  n_notes_skogsduvans_ha_utsatts_for_automatiserade_skannningar_fran_internet["Skogsduvans HA utsätts för automatiserade skannningar från internet"]
  n_notes_slutverifiering_av_sammanslagen_plattform_mal_540_genom_hela_kedjan["Slutverifiering av sammanslagen plattform: mål #540 genom hela kedjan"]
  n_notes_snabbtillagg_med_mal_och_valbart_foraldramal["Snabbtillägg med mål och valbart föräldramål"]
  n_notes_snabbtillaggets_popup_fick_escape_och_visuell_polish["Snabbtilläggets popup fick Escape och visuell polish"]
  n_notes_sortering_av_grundmal_pa_goals_med_piltangenterna["Sortering av grundmål på /goals med piltangenterna"]
  n_notes_sortering_av_undermal_med_piltangenterna_var_trasig_i_plan["Sortering av undermål med piltangenterna var trasig i plan"]
  n_notes_soulmate_teamet_i_sundsvall_importerades["Soulmate-teamet i Sundsvall importerades"]
  n_notes_sprint_timer_saknas["Sprint-läget saknar timer"]
  n_notes_stabiliserat_vertikalt_viktlage["Stabiliserat vertikalt viktläge"]
  n_notes_standardkalender_i_plan["Standardkalender i Plan"]
  n_notes_started_finished_kunde_inte_rensas_i_plan_null_ignorerades_i_go_api_t["started/finished kunde inte rensas i plan (null ignorerades i Go API:t)"]
  n_notes_stk_dev_miljon_kor_mot_neon_istallet_for_cnpg["stk: dev-miljön kör mot Neon istället för CNPG"]
  n_notes_stk_matchi_synk_sjalvaterstallande_synk_och_var_database_url_satts["STK Matchi-synk: självåterställande synk och var DATABASE_URL sätts"]
  n_notes_stk_och_sundsvall_padelklubb_styrelsemotesschema_2026_och_var_handlingarna_finns["STK och Sundsvall Padelklubb - styrelsemötesschema 2026 och var handlingarna finns"]
  n_notes_stow_fallgropar_dotfiles_i_repos_dotfiles_stow_path_repos_dotfiles["Stow-fallgropar: dotfiles i ~/repos/dotfiles (stow_path = repos/dotfiles)"]
  n_notes_stray_wiki_i_typetrainer_repot_ar_inte_kallan_for_wikin["Stray wiki/ i typetrainer-repot är inte källan för wikin"]
  n_notes_synk_hallbokningar_kommunens_actor_smartbook_matchi["Synk hallbokningar: kommunens Actor Smartbook -> matchi"]
  n_notes_taggar_pa_mal_i_plan_tangenten_g["Taggar på mål i plan - tangenten g"]
  n_notes_taggar_via_leader_t_pa_alla_malsidor["Taggar via leader+t på alla målsidor"]
  n_notes_talos_dns_fallback_mot_publika_resolvrar["Talos DNS-fallback mot publika resolvrar"]
  n_notes_talos_etcd_en_gammal_ephemeral_kopia_pa_controlplanet_ger_en_trasig_raft_log["Talos etcd: en gammal EPHEMERAL-kopia pa controlplanet ger en trasig raft-log"]
  n_notes_talos_hostname_auto_i_en_generisk_config_byter_nodens_namn_och_kastar_ut_den_ur_klustret["Talos: hostname: auto i en generisk config byter nodens namn och kastar ut den ur klustret"]
  n_notes_talos_image_cache_kraver_ephemeral_cirkulart_beroende_med_state_meta_pa_en_farsk_metal_nod["Talos image-cache kräver EPHEMERAL -> cirkulärt beroende med STATE/META på en färsk metal-nod"]
  n_notes_talos_installen_nollstaller_de_forsta_11_mb_pa_microsd_loadern_maste_aterstallas["Talos-installen nollstaller de forsta 11 MB pa microSD - loadern maste aterstallas"]
  n_notes_talos_noden_kan_hamna_i_maintenance_mode_om_state_volymen_hittas_efter_config_fasen["Talos: noden kan hamna i maintenance mode om STATE-volymen hittas efter config-fasen"]
  n_notes_talos_v1_14_machineconfig_specen_ar_en_quoted_string_och_flera_dokument["Talos v1.14: machineconfig-specen är en quoted string och flera dokument"]
  n_notes_talos_v1_14_utan_unattendedinstallconfig_oinstallerad_metal_nod_bootar_om_i_en_loop["Talos v1.14 utan UnattendedInstallConfig: oinstallerad metal-nod bootar om i en loop"]
  n_notes_talos_varningar_pa_controlplanes_90_kommer_fra_talos_discovery_service_dns["Talos-varningar pa controlplanes: 90% kommer fra Talos Discovery Service + DNS"]
  n_notes_tidsatgang_pa_mal_i_plan_estimated_minutes["Tidsåtgång på mål i Plan (estimated_minutes)"]
  n_notes_tilt_filbevakning_kan_do_starta_om_tilt_up["Tilt-filbevakning kan dö - starta om tilt up"]
  n_notes_transienta_502_fra_go_api_t_nar_poddar_startar_om_i_dev_klustret["Transienta 502 fra Go API:t nar poddar startar om i dev-klustret"]
  n_notes_tre_satt_att_pxe_boota_rock_4c_och_varfor_dhcp_metoden_valdes["Tre satt att PXE-boota ROCK 4C+ och varfor dhcp-metoden valdes"]
  n_notes_twingate_anslutaren_for_10_10_10_0_24_var_nere_recovery_floden_kan_inte_koras_fran_arbetsstationen["Twingate-anslutaren for 10.10.10.0/24 var nere - recovery-floden kan inte koras fran arbetsstationen"]
  n_notes_twingate_ressurserna_forsvann_pa_arbetsstationen_2026_11_04_hemnatet_10_1_1_1_och_datacenter_10_10_10_0_24_oppnarbara["Twingate-ressurserna forsvann pa arbetsstationen 2026-11-04 - hemnatet (10.1.1.1) och datacenter (10.10.10.0/24) oppnarbara"]
  n_notes_tydligare_hjalpmodal_for_kalenders_kortkommandon["Tydligare hjälpmodal för kalenders kortkommandon"]
  n_notes_u_boot_2026_01_rock_4c_saknar_bootfile_boot_script_dhcp_natverksboot_dor_direkt["U-Boot 2026.01 (ROCK 4C+) saknar bootfile/boot_script_dhcp - nätverksboot dör direkt"]
  n_notes_u_boot_for_rock_4c_built_from_mainline_with_pxe_first["U-Boot for ROCK 4C+ built from mainline with PXE first"]
  n_notes_u_boot_hamtar_alltid_boot_script_dhcp_per_nod_config_via_ipaddr["U-Boot hämtar alltid boot_script_dhcp - per-nod-config via ${ipaddr}"]
  n_notes_u_boot_kors_en_gang_boot_targets_styr_steg_2_inte_steg_1["U-Boot körs EN gång: boot_targets styr steg 2, inte steg 1"]
  n_notes_u7_lite_94_2a_6f_7a_4b_33_adopterar_inte_saknar_management_ip_och_nattverkssegment_pa_switchen["U7 Lite (94:2a:6f:7a:4b:33) adopterar inte - saknar management-IP och nattverkssegment pa switchen"]
  n_notes_undo_stapel_i_plan_z_angrar_senaste_malhandling["Undo-stapel i plan: Z ångrar senaste målhandling"]
  n_notes_unifi_gatewayen_har_ingen_publik_ipv6_cgnat_via_starlink["UniFi-gatewayen har ingen publik IPv6 (CGNAT via Starlink)"]
  n_notes_usb_disken_acer_wd_320gb_status_och_longhorn_bedomning["USB-disken 'ACER' (WD 320GB): status och Longhorn-bedömning"]
  n_notes_utbyggnad_av_pappas_personlista_bolagsregister_konferensbios_gav_140_namngivna_personer["Utbyggnad av pappas personlista: bolagsregister + konferensbios gav 140 namngivna personer"]
  n_notes_uttunnade_serier_vags_sa_att_antalen_blir_exakta["Uttunnade serier vägs så att antalen blir exakta"]
  n_notes_utvecklingsflode["Utvecklingsflöde"]
  n_notes_valfri_fargkod_pa_kalendrar["Valfri färgkod på kalendrar"]
  n_notes_verifierade_policy_hal_i_tool_lagret_kommandon_ar_blocklist_inte_allowlist["Verifierade policy-hål i tool-lagret - kommandon är blocklist, inte allowlist"]
  n_notes_vertikal_kalender_med_ikonstod_i_plan["Vertikal kalender med ikonstöd i Plan"]
  n_notes_vertikal_kalender_med_kontinuerlig_zoom["Vertikal kalender med kontinuerlig zoom"]
  n_notes_vertikalt_viktlage_med_j_k_h_l["Vertikalt viktläge med j k h l"]
  n_notes_viktning_av_barnmal_med_balanserad_draglinje["Viktning av barnmål med balanserad draglinje"]
  n_notes_viktning_flyttad_till_varje_mal["Viktning flyttad till varje mål"]
  n_notes_viktning_som_proportional_bar_under_maltitel["Viktning som proportional bar under måltitel"]
  n_notes_visionite_och_kvadrat_i_sundsvall_kartlagt_via_bolagsregister_visionites_kontaktsida_och_linkedin["Visionite och Kvadrat i Sundsvall - kartlagt via bolagsregister, Visionites kontaktsida och LinkedIn"]
  n_notes_vy_for_avslutade_mal["Vy för avslutade mål"]
  n_notes_windows_1252_avkodning_av_mejlens_amnesrad["Windows-1252-avkodning av mejlens ämnesrad"]
  n_notes_wordlist_kallor["Ordlistan finns på två ställen"]
  n_notes_worker_1_kortet_flashat_med_ny_mainline_build_tf_a_fran_repot["worker-1-kortet flashat med ny mainline-build (TF-A fran repot)"]
  n_notes_worker_1_och_worker_2_ar_doda_pa_link_niva_pxe_kedjan_ar_inte_blockerande["worker-1 och worker-2 ar doda pa link-niva - PXE-kedjan ar inte blockerande"]
  n_notes_worker_1_rock_4c_nar_pxe_via_u_boot_boot_targets_men_pxe_ligger_sist["worker-1 (Rock 4C+) når PXE via U-Boot boot_targets, men PXE ligger sist"]
  n_notes_worker_1_rock_4c_tyst_mainline_u_boot_byggd_men_boarden_laser_inte_kortet["worker-1 (ROCK 4C+) tyst: mainline-U-Boot byggd, men boarden läser inte kortet"]
  n_notes_worker_1_startar_om_med_strom_men_skickar_noll_paket_fysisk_omstart_bevisar_samma_fel["worker-1 startar om med stram men skickar noll paket - fysisk omstart bevisar samma fel"]
  n_notes_worker_1_svarar_inte_ens_med_en_full_poe_cykel_noll_paket_i_bade_riktningar["worker-1 svarar inte ens med en full PoE-cykel - noll paket i bade riktningar"]
  n_notes_worker_1_svarar_inte_i_natet_felsokningsstatus["worker-1 svarar inte i nätet - felsökningsstatus"]
  n_notes_worker_2_utan_strom_port_22_up_false_poe_power_0["worker-2 utan strom - port 22 up=false poe_power=0"]
  n_notes_worker_3_bootade_via_pxe_men_talos_installen_nollstallde_de_forsta_11_mb_pa_microsd["worker-3 bootade via PXE - men Talos-installen nollstallde de forsta 11 MB pa microSD"]
  n_notes_worker_3_poe_omstart_och_talos_konfiguration_2026_11_20["Worker-3 PoE-omstart och Talos-konfiguration 2026-11-20"]
  n_notes_worker_4_ateransluten_till_cluster1_via_lokalt_sd_kort["worker-4 återansluten till cluster1 via lokalt SD-kort"]
  n_notes_worker_4_rock_4se_talos_bootar_lokal_nvme_installation_apid_startar_aldrig_pxe_kedjan_redo_bootordningen_kravs_fysiskt["worker-4 (ROCK 4SE): Talos bootar lokal NVMe-installation, apid startar aldrig - PXE-kedjan redo, bootordningen kravs fysiskt"]
  n_notes_worker_4_utan_pxe_talos_installation_pa_microsd_kortet_via_image_factory_grub_layout["worker-4 utan PXE: Talos-installation pa microSD-kortet via Image Factory (GRUB-layout)"]
  n_notes_worker_7_s_kubeletcertifikat_signerades_av_en_gammal_kubernetes_ca_som_finns_kvar_i_git_historiken["Worker-7:s kubeletcertifikat signerades av en gammal Kubernetes-CA som finns kvar i git-historiken"]
  n_notes_worker_8_korrupt_xfs_pa_ephemeral_partitionen_gjorde_att_kubelet_aldrig_startade_fix_med_mkfs_xfs["worker-8: korrupt XFS pa EPHEMERAL-partitionen gjorde att kubelet aldrig startade - fix med mkfs.xfs"]
  n_notes_workers_bootar_inte_via_pxe_trots_fungerande_pxe_server_unifi_poe_styrning_via_rest_device["Workers bootar inte via PXE trots fungerande PXE-server - UniFi PoE-styrning via rest/device"]
  n_notes_workmux_merge_misslyckas_i_huvud_worktree_kora_kommandot_i_feature_worktreet["workmux merge misslyckas i huvud-worktree - kora kommandot i feature-worktreet"]
  n_summaries_16x1_hdmi_kvm_rs232_lan_protokoll["16x1 HDMI KVM - RS232/LAN-protokoll"]
  n_summaries_2026_11_03_learn_kunskapsbasen_kan_lagras_i_postgresql["'Sammanfattning: Learn-kunskapsbasen kan lagras i PostgreSQL'"]
  n_summaries_flux_components["Sammanfattning: Flux-komponenter"]
  n_summaries_infrastructure_flux["Sammanfattning: Infrastructure Flux"]
  n_summaries_kalenderns_detaljnivaer_och_lane_packning["Kalenderns detaljnivåer och lane-packning"]
  n_summaries_llm_wiki_karpathy["Sammanfattning: LLM Wiki (Karpathy)"]
  n_summaries_longhorn_disk_guide["Sammanfattning: Diskar i Longhorn"]
  n_concepts_authentik --> n_concepts_oidc
  n_concepts_authentik --> n_concepts_zitadel
  n_concepts_authentik --> n_concepts_cloudnative_pg
  n_concepts_authentik --> n_entities_kubernetes
  n_concepts_cert_manager --> n_concepts_gitops
  n_concepts_cert_manager --> n_entities_kubernetes
  n_concepts_cert_manager --> n_entities_talos_linux
  n_concepts_cloudflare_tunnel --> n_concepts_http_routing_flow
  n_concepts_cloudflare_tunnel --> n_entities_kubernetes
  n_concepts_cloudflare_tunnel --> n_concepts_sops
  n_concepts_cloudnative_pg --> n_concepts_zitadel
  n_concepts_cloudnative_pg --> n_entities_kubernetes
  n_concepts_cloudnative_pg --> n_concepts_longhorn
  n_concepts_envoy_gateway --> n_concepts_http_routing_flow
  n_concepts_envoy_gateway --> n_concepts_oidc
  n_concepts_envoy_gateway --> n_concepts_zitadel
  n_concepts_envoy_gateway --> n_entities_kubernetes
  n_concepts_envoy_gateway --> n_concepts_metallb
  n_concepts_envoy_gateway --> n_concepts_cert_manager
  n_concepts_flux_operator --> n_entities_kubernetes
  n_concepts_flux_operator --> n_summaries_flux_components
  n_concepts_flux_operator --> n_summaries_infrastructure_flux
  n_concepts_flux_operator --> n_concepts_sops
  n_concepts_flux_operator --> n_concepts_gitops
  n_concepts_flux_operator --> n_entities_flux_gitops
  n_concepts_gitops --> n_entities_kubernetes
  n_concepts_gitops --> n_summaries_infrastructure_flux
  n_concepts_gitops --> n_entities_flux_gitops
  n_concepts_gitops --> n_concepts_flux_operator
  n_concepts_gitops --> n_concepts_sops
  n_concepts_http_routing_flow --> n_concepts_cert_manager
  n_concepts_http_routing_flow --> n_concepts_envoy_gateway
  n_concepts_http_routing_flow --> n_concepts_metallb
  n_concepts_http_routing_flow --> n_concepts_cloudflare_tunnel
  n_concepts_kube_prometheus --> n_entities_kubernetes
  n_concepts_kube_prometheus --> n_concepts_envoy_gateway
  n_concepts_kubelet_csr_approver --> n_entities_kubernetes
  n_concepts_kubelet_csr_approver --> n_entities_talos_linux
  n_concepts_kyverno_cel --> n_concepts_kyverno_policies
  n_concepts_kyverno_cel --> n_entities_kubernetes
  n_concepts_kyverno_policies --> n_concepts_kyverno_cel
  n_concepts_kyverno_policies --> n_entities_kubernetes
  n_concepts_kyverno_policies --> n_concepts_flux_operator
  n_concepts_longhorn --> n_notes_usb_disken_acer_wd_320gb_status_och_longhorn_bedomning
  n_concepts_longhorn --> n_entities_talos_linux
  n_concepts_longhorn --> n_entities_kubernetes
  n_concepts_longhorn --> n_summaries_longhorn_disk_guide
  n_concepts_metallb --> n_entities_kubernetes
  n_concepts_metallb --> n_concepts_envoy_gateway
  n_concepts_oidc --> n_concepts_zitadel
  n_concepts_oidc --> n_concepts_authentik
  n_concepts_oidc --> n_concepts_envoy_gateway
  n_concepts_oidc --> n_entities_kubernetes
  n_concepts_poolad_eller_direkt_anslutning_mot_neon --> n_notes_masterdatabas_i_postgres_schema_import_och_drift_for_it_rekryteringen
  n_concepts_sops --> n_entities_kubernetes
  n_concepts_sops --> n_entities_talos_linux
  n_concepts_sops --> n_notes_klustret_anvander_sops_via_flux_inte_vault_eso
  n_concepts_sops --> n_concepts_flux_operator
  n_concepts_sops --> n_entities_flux_gitops
  n_concepts_zitadel --> n_concepts_oidc
  n_concepts_zitadel --> n_concepts_authentik
  n_concepts_zitadel --> n_concepts_envoy_gateway
  n_concepts_zitadel --> n_concepts_cloudnative_pg
  n_concepts_zitadel --> n_entities_kubernetes
  n_entities_flux_gitops --> n_entities_kubernetes
  n_entities_flux_gitops --> n_summaries_flux_components
  n_entities_flux_gitops --> n_concepts_flux_operator
  n_entities_flux_gitops --> n_concepts_gitops
  n_entities_flux_gitops --> n_concepts_sops
  n_entities_flux_gitops --> n_entities_talos_linux
  n_entities_hdmi_kvm_switch_16_port_labbet --> n_notes_kvm_styrning_fran_pxe_pi_n_16_port_hdmi_kvm
  n_entities_hdmi_kvm_switch_16_port_labbet --> n_summaries_16x1_hdmi_kvm_rs232_lan_protokoll
  n_entities_kubernetes --> n_concepts_kubelet_csr_approver
  n_entities_kubernetes --> n_concepts_oidc
  n_entities_kubernetes --> n_concepts_sops
  n_entities_kubernetes --> n_concepts_zitadel
  n_entities_kubernetes --> n_entities_pxe_server
  n_entities_kubernetes --> n_summaries_flux_components
  n_entities_kubernetes --> n_notes_klustret_har_byggts_om_bara_kube_system_flux_och_appar_saknas_2026_11_23
  n_entities_kubernetes --> n_concepts_flux_operator
  n_entities_kubernetes --> n_concepts_cert_manager
  n_entities_kubernetes --> n_concepts_longhorn
  n_entities_kubernetes --> n_concepts_metallb
  n_entities_kubernetes --> n_concepts_kube_prometheus
  n_entities_kubernetes --> n_concepts_envoy_gateway
  n_entities_kubernetes --> n_concepts_cloudnative_pg
  n_entities_kubernetes --> n_concepts_kyverno_policies
  n_entities_kubernetes --> n_notes_klustret_anvander_sops_via_flux_inte_vault_eso
  n_entities_kubernetes --> n_concepts_gitops
  n_entities_kubernetes --> n_entities_flux_gitops
  n_entities_kubernetes --> n_entities_talos_linux
  n_entities_pxe_server --> n_notes_pxe_bootloop_initramfs_startar_om_for_att_boota_installerad_talos_pa_emmc
  n_entities_pxe_server --> n_notes_workers_bootar_inte_via_pxe_trots_fungerande_pxe_server_unifi_poe_styrning_via_rest_device
  n_entities_pxe_server --> n_entities_talos_linux
  n_entities_pxe_server --> n_entities_kubernetes
  n_entities_pxe_server --> n_notes_pa_rock_plattorna_finns_ingen_bootloader_i_hardvaran
  n_entities_pxe_server --> n_notes_rock_4c_och_4se_har_ingen_spi_flash_i_fabrik_u_boot_ligger_pa_microsd_emmc
  n_entities_pxe_server --> n_notes_rk3399_s_bootrom_soker_spi_emmc_sd_usb_hardkodat
  n_entities_pxe_server --> n_notes_u_boot_kors_en_gang_boot_targets_styr_steg_2_inte_steg_1
  n_entities_pxe_server --> n_notes_u_boot_hamtar_alltid_boot_script_dhcp_per_nod_config_via_ipaddr
  n_entities_pxe_server --> n_notes_u_boot_2026_01_rock_4c_saknar_bootfile_boot_script_dhcp_natverksboot_dor_direkt
  n_entities_pxe_server --> n_notes_u_boot_for_rock_4c_built_from_mainline_with_pxe_first
  n_entities_pxe_server --> n_notes_pxe_kedjan_dhcp_boot_pekar_pa_textskript_matchbox_efi_ar_uefi_vagen
  n_entities_pxe_server --> n_notes_pxe_boot_script_dhcp_kraver_mkimage_uimage_inte_ren_hush_text
  n_entities_pxe_server --> n_notes_pxe_mainline_u_boots_bootefi_kraver_addr_size_for_image_och_initrd
  n_entities_pxe_server --> n_notes_pxe_kedjan_for_worker_1_verifierad_serverside_pxe_metoden_failar_snabbt_utan_pxelinux_cfg
  n_entities_pxe_server --> n_notes_pxe_varden_korde_foraldrad_matchbox_nodconfiger_saknades_404
  n_entities_pxe_server --> n_notes_pxe_aterstallning_nodspecifika_matchbox_profiler_och_maintenance_mode_flode
  n_entities_pxe_server --> n_notes_pxe_genererade_nodconfigs_var_stale_efter_talos_v1_14_uppgraderingen
  n_entities_pxe_server --> n_notes_gotcha_pxe_forst_gor_noden_till_en_pxe_nod_maintenance_mode
  n_entities_pxe_server --> n_notes_pxe_forst_bootloader_for_rock_4c_byggd_som_patchad_u_boot_image
  n_entities_pxe_server --> n_notes_korrigering_rock4c_plus_bin_ar_ett_microsd_emmc_image_inte_en_spi_image
  n_entities_pxe_server --> n_notes_controlplane_recovery_via_pxe_ar_inte_konfigurerad
  n_entities_pxe_server --> n_notes_dnsmasq_dhcp_boot_tar_bara_en_tagg_anvand_tag_if_for_villkor
  n_entities_pxe_server --> n_notes_tre_satt_att_pxe_boota_rock_4c_och_varfor_dhcp_metoden_valdes
  n_entities_pxe_server --> n_notes_rock_pi_4_rk3399_defconfig_bygger_en_rock_pi_4a_for_4c_ska_rock_4c_plus_rk3399_defconfig_anvandas
  n_entities_pxe_server --> n_notes_mellanslag_i_radxa_rock_4c_saboterar_byggen_losning_bind_mount
  n_entities_pxe_server --> n_notes_worker_1_kortet_flashat_med_ny_mainline_build_tf_a_fran_repot
  n_entities_pxe_server --> n_notes_worker_1_rock_4c_tyst_mainline_u_boot_byggd_men_boarden_laser_inte_kortet
  n_entities_pxe_server --> n_notes_worker_1_rock_4c_nar_pxe_via_u_boot_boot_targets_men_pxe_ligger_sist
  n_entities_pxe_server --> n_notes_worker_1_och_worker_2_ar_doda_pa_link_niva_pxe_kedjan_ar_inte_blockerande
  n_entities_pxe_server --> n_notes_worker_1_svarar_inte_i_natet_felsokningsstatus
  n_entities_pxe_server --> n_notes_worker_2_utan_strom_port_22_up_false_poe_power_0
  n_entities_pxe_server --> n_notes_simon_pxe_flash_flashar_pxe_forst_u_boot_till_sd_kort_i_pxe_servern
  n_entities_pxe_server --> n_notes_simon_pxe_flash_clean_code_omgang_2_fokuserade_funktioner
  n_entities_pxe_server --> n_notes_bash_fallgropar_i_pxe_flash_tab_som_ifs_whitespace_och_read_med_for_fa_variabler
  n_entities_talos_linux --> n_entities_kubernetes
  n_entities_talos_linux --> n_entities_pxe_server
  n_entities_talos_linux --> n_notes_usb_disken_acer_wd_320gb_status_och_longhorn_bedomning
  n_entities_talos_linux --> n_notes_talos_v1_14_machineconfig_specen_ar_en_quoted_string_och_flera_dokument
  n_entities_talos_linux --> n_concepts_flux_operator
  n_entities_talos_linux --> n_concepts_gitops
  n_entities_talos_linux --> n_concepts_sops
  n_entities_talos_linux --> n_entities_flux_gitops
  n_entities_talos_linux --> n_summaries_longhorn_disk_guide
  n_notes_1password_mcp_environments_pa_arch_omarchy_kraver_setgid_pa_1password_mcp --> n_notes_losning_1password_backed_env_variabel_for_mcp_tokens
  n_notes_accuracy_fel_omstart --> n_notes_median_i_stallet_for_snitt_for_wpm
  n_notes_accuracy_fel_omstart --> n_notes_session_livscykel
  n_notes_accuracy_fel_omstart --> n_notes_session_livscykel
  n_notes_agent_api_crashade_i_tilt_saknad_database_url_mot_cnpg_databasen_i_agents_dev --> n_notes_agent_granssnittet_nuxt_saknades_helt_i_tilt_routes_forsvann_i_refaktoreringen_677b876
  n_notes_agent_granssnittet_nuxt_saknades_helt_i_tilt_routes_forsvann_i_refaktoreringen_677b876 --> n_notes_agent_api_crashade_i_tilt_saknad_database_url_mot_cnpg_databasen_i_agents_dev
  n_notes_arbetsstationen_ar_omarchy_arch_apt_get_finns_inte_anvand_pacman_yay --> n_notes_kanata_switch_session_tmuxinator_fonster_oppnades_inte
  n_notes_arbetsstationen_ar_omarchy_arch_apt_get_finns_inte_anvand_pacman_yay --> n_notes_pi_tog_bort_pi_rate_limit_och_pi_pattern_retry_fixade_omarchy_temasy
  n_notes_arkitektur --> n_notes_bigram_algoritmen
  n_notes_arkitektur --> n_notes_session_livscykel
  n_notes_arkitektur --> n_notes_utvecklingsflode
  n_notes_arkitektur --> n_notes_kunskaps_wiki
  n_notes_arkitektur --> n_notes_session_livscykel
  n_notes_arkitektur --> n_notes_bigram_algoritmen
  n_notes_artikeln_om_agentic_software_factory_vad_som_ar_relevant_for_agentsystemet --> n_notes_verifierade_policy_hal_i_tool_lagret_kommandon_ar_blocklist_inte_allowlist
  n_notes_aterkommande_serier_doljs_inte_langre_beroende_pa_zoomniva --> n_notes_en_uttunnad_forekomst_maste_vara_synlig_aven_om_den_ligger_utanfor_fonstret
  n_notes_authflytten_i_plan_sessionen_ligger_nu_i_go_api_t_inte_i_nuxt --> n_concepts_zitadel
  n_notes_authflytten_i_plan_sessionen_ligger_nu_i_go_api_t_inte_i_nuxt --> n_notes_dev_inloggning_i_plan_session_token_db_rattigheter_och_oauth_url_er
  n_notes_authflytten_i_plan_sessionen_ligger_nu_i_go_api_t_inte_i_nuxt --> n_notes_started_finished_kunde_inte_rensas_i_plan_null_ignorerades_i_go_api_t
  n_notes_barnordning_foljer_beroenden_i_malsidan --> n_notes_sortering_av_undermal_med_piltangenterna_var_trasig_i_plan
  n_notes_bash_fallgropar_i_pxe_flash_tab_som_ifs_whitespace_och_read_med_for_fa_variabler --> n_entities_pxe_server
  n_notes_beslut_pi_durable_satts_pa_hyllan_for_tillfallet --> n_notes_md_viewer_tauri_installerades_men_avinstallerades_igen
  n_notes_beslut_pi_durable_satts_pa_hyllan_for_tillfallet --> n_notes_pi_desktop_notis_nar_agent_ar_klar
  n_notes_beslut_pi_durable_satts_pa_hyllan_for_tillfallet --> n_notes_pi_durable_1_4_0_api_avvikelser_fran_blogginlagget_och_tva_gotchas
  n_notes_beslut_pi_durable_satts_pa_hyllan_for_tillfallet --> n_notes_pi_durable_demo_paketet_earendil_works_pi_durable_fungerar
  n_notes_beslut_pi_durable_satts_pa_hyllan_for_tillfallet --> n_notes_pi_durable_fungerar_med_minimax_coding_plan_anthropic_messages_api
  n_notes_bigram_algoritmen --> n_notes_arkitektur
  n_notes_bigram_algoritmen --> n_notes_wordlist_kallor
  n_notes_bigram_algoritmen --> n_notes_arkitektur
  n_notes_cachade_imap_sessioner_kraver_hantering_av_doda_anslutningar --> n_notes_mailmaster_kontodropdown_outlook_inloggning_via_device_code
  n_notes_cluster1_stale_kubeconfig_ca_och_noll_workers_diagnos_2026_11_11 --> n_notes_workers_bootar_inte_via_pxe_trots_fungerande_pxe_server_unifi_poe_styrning_via_rest_device
  n_notes_cluster1_stale_kubeconfig_ca_och_noll_workers_diagnos_2026_11_11 --> n_notes_worker_8_korrupt_xfs_pa_ephemeral_partitionen_gjorde_att_kubelet_aldrig_startade_fix_med_mkfs_xfs
  n_notes_controlplane_recovery_via_pxe_ar_inte_konfigurerad --> n_entities_pxe_server
  n_notes_controlplane_recovery_via_pxe_ar_inte_konfigurerad --> n_notes_talos_etcd_en_gammal_ephemeral_kopia_pa_controlplanet_ger_en_trasig_raft_log
  n_notes_cookies_json_var_versionsstyrd_i_offentligt_neuro_repo --> n_notes_neon_api_token_hamnade_i_dotfiles_git_historiken
  n_notes_cookies_json_var_versionsstyrd_i_offentligt_neuro_repo --> n_notes_neuroquant_vikterna_raknas_om_dagligen_mejl_triggas_pa_publicering
  n_notes_detaljvy_for_kandidat_i_pappa_gemensamma_vagar_mellan_kandidater --> n_notes_redigerare_och_agarskap_for_relationer_i_pappa
  n_notes_detaljvy_for_kandidat_i_pappa_gemensamma_vagar_mellan_kandidater --> n_notes_relationer_mellan_kandidater_i_pappa_symmetriska_par_med_styrka_1_10
  n_notes_dnsmasq_dhcp_boot_tar_bara_en_tagg_anvand_tag_if_for_villkor --> n_entities_pxe_server
  n_notes_en_uttunnad_forekomst_maste_vara_synlig_aven_om_den_ligger_utanfor_fonstret --> n_notes_uttunnade_serier_vags_sa_att_antalen_blir_exakta
  n_notes_enter_pa_goals_skapar_nytt_grundmal_istallet_for_att_oppna_malet --> n_notes_eviga_mal_i_plan_kan_inte_slutforas
  n_notes_enter_pa_goals_skapar_nytt_grundmal_istallet_for_att_oppna_malet --> n_notes_goals_listar_bara_grundmal_scroll_vid_j_k
  n_notes_enter_pa_goals_skapar_nytt_grundmal_istallet_for_att_oppna_malet --> n_notes_sortering_av_grundmal_pa_goals_med_piltangenterna
  n_notes_erfarenhetsfilter_i_pappa_it_experience_years_med_sparbart_grund --> n_notes_masterdatabas_i_postgres_schema_import_och_drift_for_it_rekryteringen
  n_notes_erfarenhetsfilter_i_pappa_it_experience_years_med_sparbart_grund --> n_notes_nuxt_4_webbapp_for_pappa_laser_personerna_ur_neon
  n_notes_erfarenhetsfilter_i_pappa_it_experience_years_med_sparbart_grund --> n_notes_pappa_anstallningshistorik_i_cv_format_fix_av_kompetensordboken
  n_notes_eviga_mal_i_plan_kan_inte_slutforas --> n_notes_mal_kan_ha_due_datum_deadline
  n_notes_eviga_mal_i_plan_kan_inte_slutforas --> n_notes_mal_kan_ha_en_lang_beskrivning_redigerad_med_tangenten_t
  n_notes_eviga_mal_i_plan_kan_inte_slutforas --> n_notes_manuell_sortering_av_mal_pa_priority
  n_notes_eviga_mal_i_plan_kan_inte_slutforas --> n_notes_priorityfilter_ska_behandla_saknat_started_som_ej_paborjat
  n_notes_expanderade_eml_laggs_in_som_riktiga_mejl_i_mappen_append --> n_notes_mailagenten_expanderar_bifogade_eml_till_egna_mejl
  n_notes_expanderade_eml_laggs_in_som_riktiga_mejl_i_mappen_append --> n_notes_moderposten_arkiveras_nar_eml_bilagorna_har_lagts_in
  n_notes_flera_kalendrar_med_egna_farger_i_plan --> n_notes_handelsen_har_ingen_egen_farg
  n_notes_flera_kalendrar_med_egna_farger_i_plan --> n_notes_standardkalender_i_plan
  n_notes_flera_kalendrar_med_egna_farger_i_plan --> n_notes_valfri_fargkod_pa_kalendrar
  n_notes_flera_kalendrar_med_egna_farger_i_plan --> n_notes_vertikal_kalender_med_ikonstod_i_plan
  n_notes_go_integrationstest_tomde_kalenderhandelserna_i_neon_igen --> n_notes_kalenderhandelser_aterstallda_fran_neon_wal_historik
  n_notes_go_integrationstest_tomde_kalenderhandelserna_i_neon_igen --> n_notes_parallella_agent_sessioner_i_samma_repo_och_databas
  n_notes_gotcha_pxe_forst_gor_noden_till_en_pxe_nod_maintenance_mode --> n_entities_pxe_server
  n_notes_gotcha_pxe_forst_gor_noden_till_en_pxe_nod_maintenance_mode --> n_notes_pxe_forst_bootloader_for_rock_4c_byggd_som_patchad_u_boot_image
  n_notes_gotcha_pxe_forst_gor_noden_till_en_pxe_nod_maintenance_mode --> n_notes_talos_image_cache_kraver_ephemeral_cirkulart_beroende_med_state_meta_pa_en_farsk_metal_nod
  n_notes_gotcha_pxe_forst_gor_noden_till_en_pxe_nod_maintenance_mode --> n_notes_talos_noden_kan_hamna_i_maintenance_mode_om_state_volymen_hittas_efter_config_fasen
  n_notes_gotcha_pxe_forst_gor_noden_till_en_pxe_nod_maintenance_mode --> n_notes_u_boot_kors_en_gang_boot_targets_styr_steg_2_inte_steg_1
  n_notes_gotcha_pxe_forst_gor_noden_till_en_pxe_nod_maintenance_mode --> n_notes_worker_1_rock_4c_nar_pxe_via_u_boot_boot_targets_men_pxe_ligger_sist
  n_notes_ha_automationer_behover_en_startv_g_for_att_aterstalla_laget_och_tre_syntaxfallor --> n_notes_skogsduvan_temperaturkaskad_i_tre_steg_styrd_av_minsta_temperaturen
  n_notes_handelsen_har_ingen_egen_farg --> n_notes_flera_kalendrar_med_egna_farger_i_plan
  n_notes_handelsen_har_ingen_egen_farg --> n_notes_standardkalender_i_plan
  n_notes_ikonvaljaren_i_plan_api_icons_routen_var_borttagen_och_ikonprovider_none_gjorde_att_inga_ikoner_renderades --> n_notes_plan_cachad_spa_skal_gjorde_att_webblasaren_korde_gammalt_js_bundle_med_ikon_provider_none
  n_notes_imap_inloggning_for_michael_brundin_soulmate_se_kraver_admin_consent --> n_notes_mailmaster_kontodropdown_outlook_inloggning_via_device_code
  n_notes_imap_inloggning_for_michael_brundin_soulmate_se_kraver_admin_consent --> n_notes_sales_soulmate_se_appen_kan_inte_logga_in_oauth_clienten_saknas_i_soulmate_tenanten
  n_notes_imap_inloggning_for_michael_brundin_soulmate_se_kraver_admin_consent --> n_notes_simonbrundin_gmail_com_inkorgen_kan_lasas_via_imap_begbot_losenordet
  n_notes_infrastructure_agents_md_uppdaterad_tre_fallor_for_agenter --> n_concepts_gitops
  n_notes_infrastructure_agents_md_uppdaterad_tre_fallor_for_agenter --> n_concepts_sops
  n_notes_ipv6_ingress_till_envoy_designlage_2026_11_15 --> n_notes_unifi_gatewayen_har_ingen_publik_ipv6_cgnat_via_starlink
  n_notes_ipv6_ingress_till_envoy_designlage_2026_11_15 --> n_notes_workers_bootar_inte_via_pxe_trots_fungerande_pxe_server_unifi_poe_styrning_via_rest_device
  n_notes_ipv6_ingress_till_envoy_designlage_2026_11_15 --> n_concepts_http_routing_flow
  n_notes_it_arbetsplatser_i_sundsvall_3_469_enligt_scb_yrkesregistret --> n_notes_rekrytering_av_it_personal_i_sundsvall_marknadsbild_och_frolista
  n_notes_kalenderhandelser_aterstallda_fran_neon_wal_historik --> n_notes_go_integrationstest_tomde_kalenderhandelserna_i_neon_igen
  n_notes_kalendern_begransar_aterkommande_forekomster_i_langtidvyer --> n_notes_aterkommande_serier_doljs_inte_langre_beroende_pa_zoomniva
  n_notes_kalendern_begransar_aterkommande_forekomster_i_langtidvyer --> n_notes_kalendern_hamtar_alla_handelser_en_gang_och_expanderar_serier_lokalt
  n_notes_kalenderns_detaljnivaer_fem_nivaer_och_lane_packning --> n_notes_kalenderns_nivavaxling_animeras_crossfade_pop_in_och_hysteresis
  n_notes_kalenderns_detaljnivaer_fem_nivaer_och_lane_packning --> n_notes_vertikal_kalender_med_ikonstod_i_plan
  n_notes_kalenderns_detaljnivaer_fem_nivaer_och_lane_packning --> n_notes_vertikal_kalender_med_kontinuerlig_zoom
  n_notes_kalenderns_detaljnivaer_fem_nivaer_och_lane_packning --> n_summaries_kalenderns_detaljnivaer_och_lane_packning
  n_notes_kalenderns_nivavaxling_animeras_crossfade_pop_in_och_hysteresis --> n_notes_kalenderns_detaljnivaer_fem_nivaer_och_lane_packning
  n_notes_kalenderns_nivavaxling_animeras_crossfade_pop_in_och_hysteresis --> n_notes_kalenderns_tidsaxelrader_animeras_vid_zoom
  n_notes_kalenderns_nivavaxling_animeras_crossfade_pop_in_och_hysteresis --> n_notes_vertikal_kalender_med_kontinuerlig_zoom
  n_notes_kalenderns_tidsaxelrader_animeras_vid_zoom --> n_notes_kalenderns_nivavaxling_animeras_crossfade_pop_in_och_hysteresis
  n_notes_kameo_beloppet_kommer_fran_transaktionerna_inte_fran_mejlet --> n_notes_kameo_utbetalningar_nuxt_sida_som_laser_lan_fran_kameo
  n_notes_kameo_beloppet_kommer_fran_transaktionerna_inte_fran_mejlet --> n_notes_rantebetalningsmejl_hamnar_i_inbox_deleted_messages_och_arkivet_matchades_aldrig
  n_notes_kameo_utbetalningar_nuxt_sida_som_laser_lan_fran_kameo --> n_notes_kameo_beloppet_kommer_fran_transaktionerna_inte_fran_mejlet
  n_notes_kameo_utbetalningar_nuxt_sida_som_laser_lan_fran_kameo --> n_notes_kameo_ranteutbetalning_redan_utbetald_samlingsuttag_pa_47_29_kr
  n_notes_kameo_utbetalningar_nuxt_sida_som_laser_lan_fran_kameo --> n_notes_kameo_sidan_flyttad_till_toppniva_och_delad_i_flikar
  n_notes_kameo_utbetalningar_nuxt_sida_som_laser_lan_fran_kameo --> n_notes_kameo_uppgifter_synkas_till_fil_sa_docker_slipper_op_cli
  n_notes_kandidatlistan_heter_nu_personer --> n_notes_nuxt_sidomeny_kandidater_som_startsida_i_pappa
  n_notes_konsulter_i_sundsvall_antalet_gar_inte_att_mata_det_som_gar_att_saga --> n_notes_it_arbetsplatser_i_sundsvall_3_469_enligt_scb_yrkesregistret
  n_notes_konsulter_i_sundsvall_antalet_gar_inte_att_mata_det_som_gar_att_saga --> n_notes_marknadslage_panel_hogst_upp_pa_personer_i_pappa
  n_notes_konsulter_vs_anstallda_i_sundsvall_halva_it_styrkan_sitter_i_konsultnara_bolag --> n_notes_it_arbetsplatser_i_sundsvall_3_469_enligt_scb_yrkesregistret
  n_notes_konsulter_vs_anstallda_i_sundsvall_halva_it_styrkan_sitter_i_konsultnara_bolag --> n_notes_konsulter_i_sundsvall_antalet_gar_inte_att_mata_det_som_gar_att_saga
  n_notes_konsulter_vs_anstallda_i_sundsvall_halva_it_styrkan_sitter_i_konsultnara_bolag --> n_notes_marknadslage_panel_hogst_upp_pa_personer_i_pappa
  n_notes_konventioner --> n_notes_arkitektur
  n_notes_konventioner --> n_notes_utvecklingsflode
  n_notes_konventioner --> n_notes_utvecklingsflode
  n_notes_korrigering_rock4c_plus_bin_ar_ett_microsd_emmc_image_inte_en_spi_image --> n_entities_pxe_server
  n_notes_korrigering_rock4c_plus_bin_ar_ett_microsd_emmc_image_inte_en_spi_image --> n_notes_rock_4c_och_4se_har_ingen_spi_flash_i_fabrik_u_boot_ligger_pa_microsd_emmc
  n_notes_korrigering_rock4c_plus_bin_ar_ett_microsd_emmc_image_inte_en_spi_image --> n_notes_u_boot_2026_01_rock_4c_saknar_bootfile_boot_script_dhcp_natverksboot_dor_direkt
  n_notes_korrigering_rock4c_plus_bin_ar_ett_microsd_emmc_image_inte_en_spi_image --> n_notes_u_boot_for_rock_4c_built_from_mainline_with_pxe_first
  n_notes_krypterade_secrets_i_klustret_ogiltiga_inloggningar_och_keycloak_brute_force --> n_notes_synk_hallbokningar_kommunens_actor_smartbook_matchi
  n_notes_kunskaps_wiki --> n_notes_globala_agentinstruktioner_deployas_via_dotfiles
  n_notes_kunskaps_wiki --> n_notes_infrastructure_agents_md_uppdaterad_tre_fallor_for_agenter
  n_notes_kunskaps_wiki --> n_notes_konventioner
  n_notes_kunskaps_wiki --> n_notes_kureringspasset_hittade_foraldrad_kunskap_i_wikin
  n_notes_kunskaps_wiki --> n_notes_stray_wiki_i_typetrainer_repot_ar_inte_kallan_for_wikin
  n_notes_kunskaps_wiki --> n_notes_utvecklingsflode
  n_notes_kunskaps_wiki --> n_summaries_llm_wiki_karpathy
  n_notes_kvm_styrning_fran_pxe_pi_n_16_port_hdmi_kvm --> n_entities_hdmi_kvm_switch_16_port_labbet
  n_notes_kvm_styrning_fran_pxe_pi_n_16_port_hdmi_kvm --> n_notes_picokvm_luckfox_startar_inte_i_labbet_lite_saknar_emmc_ingen_poe
  n_notes_kvm_styrning_fran_pxe_pi_n_16_port_hdmi_kvm --> n_notes_simon_kvm_cli_kommando_for_att_byta_hdmi_kvm_port_fran_pxe_vard
  n_notes_kvm_styrning_fran_pxe_pi_n_16_port_hdmi_kvm --> n_summaries_16x1_hdmi_kvm_rs232_lan_protokoll
  n_notes_leader_key_gjordes_global_mellanslag_sekvenser_pa_alla_sidor --> n_notes_mellanslag_blockerades_i_malbeskrivningen
  n_notes_learn_calibre_synk_och_automatisk_import_via_systemd --> n_notes_learn_kunskapsbasen_kan_lagras_i_postgresql
  n_notes_learn_korsanalys_av_varje_bok_mot_den_samlade_kunskapsbasen --> n_notes_learn_kunskapsbasen_kan_lagras_i_postgresql
  n_notes_learn_kunskapsbasen_kan_lagras_i_postgresql --> n_notes_learn_korsanalys_av_varje_bok_mot_den_samlade_kunskapsbasen
  n_notes_learn_kunskapsbasen_kan_lagras_i_postgresql --> n_summaries_2026_11_03_learn_kunskapsbasen_kan_lagras_i_postgresql
  n_notes_learn_kunskapssystem_implementerat_i_src_python --> n_notes_learn_calibre_synk_och_automatisk_import_via_systemd
  n_notes_learn_kunskapssystem_implementerat_i_src_python --> n_notes_learn_kunskapssystemet_portat_fran_python_till_go
  n_notes_learn_kunskapssystemet_portat_fran_python_till_go --> n_notes_learn_kunskapssystem_implementerat_i_src_python
  n_notes_learn_kunskapssystemet_portat_fran_python_till_go --> n_notes_learn_rest_api_och_nuxt_spa_mot_samma_kunskapsbas
  n_notes_learn_rest_api_och_nuxt_spa_mot_samma_kunskapsbas --> n_notes_learn_designsystem_a3_folio_i_nuxt
  n_notes_learn_rest_api_och_nuxt_spa_mot_samma_kunskapsbas --> n_notes_learn_kunskapssystemet_portat_fran_python_till_go
  n_notes_lightcraft_photocraft_installerade_pa_omarchy_workstation_utan_sudo --> n_notes_photocraft_agentstyrning_via_control_protokoll_control_7878
  n_notes_lokal_expansion_kan_verifieras_mot_serverns_expansion --> n_notes_kalendern_hamtar_alla_handelser_en_gang_och_expanderar_serier_lokalt
  n_notes_losning_1password_backed_env_variabel_for_mcp_tokens --> n_notes_neon_api_token_hamnade_i_dotfiles_git_historiken
  n_notes_mailagenten_expanderar_bifogade_eml_till_egna_mejl --> n_notes_expanderade_eml_laggs_in_som_riktiga_mejl_i_mappen_append
  n_notes_mailagenten_flyttad_fran_mailmaster_repot_in_i_pappa --> n_notes_mailagent_bygget_brot_pa_ofardig_kameo_integration
  n_notes_mailagenten_flyttad_fran_mailmaster_repot_in_i_pappa --> n_notes_mailmaster_gammal_container_build_korde_process_och_arkiverade_mejl
  n_notes_mailagenten_flyttad_fran_mailmaster_repot_in_i_pappa --> n_notes_pappa_regelendpointet_laser_och_sparar_riktig_rules_md_var_stub
  n_notes_mailagentens_kontodropdown_renderades_aldrig_nuxt_prefixade_komponenter --> n_notes_mailagentens_kontofel_gav_alltid_http_500_och_listan_foljde_inte_kontot
  n_notes_mailagentens_kontodropdown_renderades_aldrig_nuxt_prefixade_komponenter --> n_notes_mailagentens_mcp_verktygen_saknade_konto_och_hade_fel_schema
  n_notes_mailagentens_kontodropdown_renderades_aldrig_nuxt_prefixade_komponenter --> n_notes_mailmaster_api_v1_gav_failed_to_fetch_go_servern_panickade_och_saknade_cors
  n_notes_mailagentens_kontofel_gav_alltid_http_500_och_listan_foljde_inte_kontot --> n_notes_mailagentens_kontodropdown_renderades_aldrig_nuxt_prefixade_komponenter
  n_notes_mailagentens_kontofel_gav_alltid_http_500_och_listan_foljde_inte_kontot --> n_notes_mailagentens_mcp_verktygen_saknade_konto_och_hade_fel_schema
  n_notes_mailagentens_mcp_verktygen_saknade_konto_och_hade_fel_schema --> n_notes_mailagenten_flyttad_fran_mailmaster_repot_in_i_pappa
  n_notes_mailagentens_mcp_verktygen_saknade_konto_och_hade_fel_schema --> n_notes_mailagentens_kontodropdown_renderades_aldrig_nuxt_prefixade_komponenter
  n_notes_mailagentens_mcp_verktygen_saknade_konto_och_hade_fel_schema --> n_notes_mailmaster_svarmejl_fick_fel_brodtext_tre_buggar_i_svarttextskedjan
  n_notes_mailmaster_kontodropdown_outlook_inloggning_via_device_code --> n_notes_cachade_imap_sessioner_kraver_hantering_av_doda_anslutningar
  n_notes_mailmaster_kontodropdown_outlook_inloggning_via_device_code --> n_notes_imap_inloggning_for_michael_brundin_soulmate_se_kraver_admin_consent
  n_notes_mailmaster_kontodropdown_outlook_inloggning_via_device_code --> n_notes_one_com_lagrar_mappar_i_nfd_normalisera_i_getfoldername
  n_notes_mailmaster_svarmejl_fick_fel_brodtext_tre_buggar_i_svarttextskedjan --> n_notes_mailmaster_granskning_av_ai_atgarder_fore_utforande
  n_notes_mailvyn_mail_omstegad_till_tailwind_tema --> n_notes_plan_appen_morkt_designsystem_med_semantiska_farg_token
  n_notes_mal_kan_ha_due_datum_deadline --> n_notes_mal_kan_ha_startdatum_syns_i_priority_forst_nar_datumet_passerat
  n_notes_mal_kan_ha_en_lang_beskrivning_redigerad_med_tangenten_t --> n_notes_enter_pa_goals_skapar_nytt_grundmal_istallet_for_att_oppna_malet
  n_notes_mal_kan_ha_en_lang_beskrivning_redigerad_med_tangenten_t --> n_notes_eviga_mal_i_plan_kan_inte_slutforas
  n_notes_mal_kan_ha_en_lang_beskrivning_redigerad_med_tangenten_t --> n_notes_mellanslag_blockerades_i_malbeskrivningen
  n_notes_mal_kan_ha_startdatum_syns_i_priority_forst_nar_datumet_passerat --> n_notes_mal_kan_ha_due_datum_deadline
  n_notes_malgraph_navigering_mellan_malsidor_gar_fran_sekunder_till_millisekunder --> n_notes_platser_pa_mal_och_kalenderhandelser_openspec_change_add_places
  n_notes_malgraph_navigering_mellan_malsidor_gar_fran_sekunder_till_millisekunder --> n_notes_sortering_av_undermal_med_piltangenterna_var_trasig_i_plan
  n_notes_manuell_sortering_av_mal_pa_priority --> n_notes_mal_kan_ha_due_datum_deadline
  n_notes_marknadslage_panel_hogst_upp_pa_personer_i_pappa --> n_notes_it_arbetsplatser_i_sundsvall_3_469_enligt_scb_yrkesregistret
  n_notes_masterdatabas_i_postgres_schema_import_och_drift_for_it_rekryteringen --> n_notes_erfarenhetsfilter_i_pappa_it_experience_years_med_sparbart_grund
  n_notes_masterdatabas_i_postgres_schema_import_och_drift_for_it_rekryteringen --> n_notes_nuxt_4_webbapp_for_pappa_laser_personerna_ur_neon
  n_notes_masterdatabas_i_postgres_schema_import_och_drift_for_it_rekryteringen --> n_notes_pappa_anstallningshistorik_i_cv_format_fix_av_kompetensordboken
  n_notes_masterdatabas_i_postgres_schema_import_och_drift_for_it_rekryteringen --> n_notes_pappa_dubblettknapp_mansklig_granskning_och_bestandig_sammanslagning
  n_notes_masterdatabas_i_postgres_schema_import_och_drift_for_it_rekryteringen --> n_notes_pappa_kor_nu_pa_neon_neon_ts_env_local_och_poolad_direkt_wiring
  n_notes_masterdatabas_i_postgres_schema_import_och_drift_for_it_rekryteringen --> n_notes_pappa_yaml_frolistan_kan_aterskapas_ur_masterdatabasen
  n_notes_masterdatabas_i_postgres_schema_import_och_drift_for_it_rekryteringen --> n_notes_rekrytering_av_it_personal_i_sundsvall_marknadsbild_och_frolista
  n_notes_matchi_bokning_slot_id_andras_nar_bokningen_skapas --> n_notes_boka_enskild_kommunbokning_i_matchi_fran_bookings
  n_notes_matchi_bokning_slot_id_andras_nar_bokningen_skapas --> n_notes_synk_hallbokningar_kommunens_actor_smartbook_matchi
  n_notes_matchi_keycloak_session_code_maste_fangas_med_generell_teckenklass --> n_notes_matchi_inloggningen_ar_flaky_klienten_forsoker_om_och_dev_klustret_behover_ombyggnad
  n_notes_matchi_keycloak_session_code_maste_fangas_med_generell_teckenklass --> n_notes_synk_hallbokningar_kommunens_actor_smartbook_matchi
  n_notes_median_i_stallet_for_snitt_for_wpm --> n_notes_utvecklingsflode
  n_notes_mellanslag_blockerades_i_malbeskrivningen --> n_notes_leader_key_gjordes_global_mellanslag_sekvenser_pa_alla_sidor
  n_notes_mellanslag_i_radxa_rock_4c_saboterar_byggen_losning_bind_mount --> n_entities_pxe_server
  n_notes_mellanslag_i_radxa_rock_4c_saboterar_byggen_losning_bind_mount --> n_notes_rock_pi_4_rk3399_defconfig_bygger_en_rock_pi_4a_for_4c_ska_rock_4c_plus_rk3399_defconfig_anvandas
  n_notes_mellanslag_i_radxa_rock_4c_saboterar_byggen_losning_bind_mount --> n_notes_worker_1_rock_4c_tyst_mainline_u_boot_byggd_men_boarden_laser_inte_kortet
  n_notes_michael_brundin_blir_en_personnod_i_relationsgrafen --> n_notes_relationsgrafen_anvander_fast_michael_id_utan_person_kind
  n_notes_min_kontaktlista_jag_som_nod_i_grafen_pappa --> n_notes_redigerare_och_agarskap_for_relationer_i_pappa
  n_notes_moderposten_arkiveras_nar_eml_bilagorna_har_lagts_in --> n_notes_expanderade_eml_laggs_in_som_riktiga_mejl_i_mappen_append
  n_notes_neon_api_token_hamnade_i_dotfiles_git_historiken --> n_notes_losning_1password_backed_env_variabel_for_mcp_tokens
  n_notes_neuroquant_ar_ombyggt_trpc_api_och_e_post_verifiering --> n_notes_neuroquant_kontot_last_av_for_manga_inloggningsforsok
  n_notes_neuroquant_ar_ombyggt_trpc_api_och_e_post_verifiering --> n_notes_neuroquant_vikterna_raknas_om_dagligen_mejl_triggas_pa_publicering
  n_notes_neuroquant_kontot_last_av_for_manga_inloggningsforsok --> n_notes_neuroquant_ar_ombyggt_trpc_api_och_e_post_verifiering
  n_notes_neuroquant_kontot_last_av_for_manga_inloggningsforsok --> n_notes_neuroquant_vikterna_raknas_om_dagligen_mejl_triggas_pa_publicering
  n_notes_neuroquant_vikterna_raknas_om_dagligen_mejl_triggas_pa_publicering --> n_notes_cookies_json_var_versionsstyrd_i_offentligt_neuro_repo
  n_notes_neuroquant_vikterna_raknas_om_dagligen_mejl_triggas_pa_publicering --> n_notes_neuroquant_ar_ombyggt_trpc_api_och_e_post_verifiering
  n_notes_neuroquant_vikterna_raknas_om_dagligen_mejl_triggas_pa_publicering --> n_notes_neuroquant_kontot_last_av_for_manga_inloggningsforsok
  n_notes_nuxt_4_webbapp_for_pappa_laser_personerna_ur_neon --> n_notes_lunchmenyn_delar_en_modell_mellan_byggskript_och_webblasare
  n_notes_nuxt_4_webbapp_for_pappa_laser_personerna_ur_neon --> n_notes_mailvyn_mail_omstegad_till_tailwind_tema
  n_notes_nuxt_4_webbapp_for_pappa_laser_personerna_ur_neon --> n_notes_masterdatabas_i_postgres_schema_import_och_drift_for_it_rekryteringen
  n_notes_nuxt_4_webbapp_for_pappa_laser_personerna_ur_neon --> n_notes_nuxt_cli_via_bun_run_dev_gar_sonder_i_stk
  n_notes_nuxt_4_webbapp_for_pappa_laser_personerna_ur_neon --> n_notes_nuxt_sidomeny_kandidater_som_startsida_i_pappa
  n_notes_nuxt_4_webbapp_for_pappa_laser_personerna_ur_neon --> n_notes_pappas_dev_miljo_kor_mot_neon_i_stallet_for_in_kluster_postgres
  n_notes_nuxt_sidomeny_kandidater_som_startsida_i_pappa --> n_notes_kandidatlistan_heter_nu_personer
  n_notes_nya_mal_far_ikon_fran_titeln --> n_notes_automatisk_malikon_behovde_sokord_och_svenska_alias
  n_notes_nya_mal_far_ikon_fran_titeln --> n_notes_malikoner_anvander_enkel_oversattning_fore_iconify
  n_notes_pa_rock_plattorna_finns_ingen_bootloader_i_hardvaran --> n_entities_pxe_server
  n_notes_pa_rock_plattorna_finns_ingen_bootloader_i_hardvaran --> n_notes_rk3399_s_bootrom_soker_spi_emmc_sd_usb_hardkodat
  n_notes_pa_rock_plattorna_finns_ingen_bootloader_i_hardvaran --> n_notes_u_boot_2026_01_rock_4c_saknar_bootfile_boot_script_dhcp_natverksboot_dor_direkt
  n_notes_pa_rock_plattorna_finns_ingen_bootloader_i_hardvaran --> n_notes_u_boot_kors_en_gang_boot_targets_styr_steg_2_inte_steg_1
  n_notes_pappa_admin_token_ett_stalle_for_token_inte_handpallagning_per_miljo --> n_notes_admin_inloggning_via_losenord_i_stallet_for_token_i_ui_t
  n_notes_pappa_admin_token_ett_stalle_for_token_inte_handpallagning_per_miljo --> n_notes_pappa_dubblettknapp_mansklig_granskning_och_bestandig_sammanslagning
  n_notes_pappa_admin_token_ett_stalle_for_token_inte_handpallagning_per_miljo --> n_notes_pappa_hemlighet_i_build_context_env_i_docker_imagen
  n_notes_pappa_admin_token_ett_stalle_for_token_inte_handpallagning_per_miljo --> n_notes_pappa_vyn_uppdrag_kravformular_och_rankad_kandidatlista
  n_notes_pappa_anstallningshistorik_i_cv_format_fix_av_kompetensordboken --> n_notes_erfarenhetsfilter_i_pappa_it_experience_years_med_sparbart_grund
  n_notes_pappa_anstallningshistorik_i_cv_format_fix_av_kompetensordboken --> n_notes_masterdatabas_i_postgres_schema_import_och_drift_for_it_rekryteringen
  n_notes_pappa_anstallningshistorik_i_cv_format_fix_av_kompetensordboken --> n_notes_rekrytering_av_it_personal_i_sundsvall_marknadsbild_och_frolista
  n_notes_pappa_anstallningshistorik_i_cv_format_fix_av_kompetensordboken --> n_notes_utbyggnad_av_pappas_personlista_bolagsregister_konferensbios_gav_140_namngivna_personer
  n_notes_pappa_dubblettknapp_mansklig_granskning_och_bestandig_sammanslagning --> n_notes_masterdatabas_i_postgres_schema_import_och_drift_for_it_rekryteringen
  n_notes_pappa_dubblettknapp_mansklig_granskning_och_bestandig_sammanslagning --> n_notes_pappa_uppdragsmatchning_delpoang_forklarande_ranking
  n_notes_pappa_dubblettknapp_mansklig_granskning_och_bestandig_sammanslagning --> n_notes_personer_kan_laggas_in_manuellt_i_pappa
  n_notes_pappa_hemlighet_i_build_context_env_i_docker_imagen --> n_notes_pappa_admin_token_ett_stalle_for_token_inte_handpallagning_per_miljo
  n_notes_pappa_kor_nu_pa_neon_neon_ts_env_local_och_poolad_direkt_wiring --> n_concepts_neon_cli_var_neon_och_neon_ts_maste_ligga
  n_notes_pappa_kor_nu_pa_neon_neon_ts_env_local_och_poolad_direkt_wiring --> n_concepts_poolad_eller_direkt_anslutning_mot_neon
  n_notes_pappa_kor_nu_pa_neon_neon_ts_env_local_och_poolad_direkt_wiring --> n_notes_pappas_dev_miljo_kor_mot_neon_i_stallet_for_in_kluster_postgres
  n_notes_pappa_uppdragsmatchning_delpoang_forklarande_ranking --> n_notes_pappa_dubblettknapp_mansklig_granskning_och_bestandig_sammanslagning
  n_notes_pappa_uppdragsmatchning_delpoang_forklarande_ranking --> n_notes_masterdatabas_i_postgres_schema_import_och_drift_for_it_rekryteringen
  n_notes_pappa_uppdragsmatchning_delpoang_forklarande_ranking --> n_notes_rekrytering_av_it_personal_i_sundsvall_marknadsbild_och_frolista
  n_notes_pappa_vyn_uppdrag_kravformular_och_rankad_kandidatlista --> n_notes_pappa_dubblettknapp_mansklig_granskning_och_bestandig_sammanslagning
  n_notes_pappa_vyn_uppdrag_kravformular_och_rankad_kandidatlista --> n_notes_pappa_uppdragsmatchning_delpoang_forklarande_ranking
  n_notes_pappa_yaml_frolistan_kan_aterskapas_ur_masterdatabasen --> n_notes_masterdatabas_i_postgres_schema_import_och_drift_for_it_rekryteringen
  n_notes_pappa_yaml_frolistan_kan_aterskapas_ur_masterdatabasen --> n_notes_pappa_anstallningshistorik_i_cv_format_fix_av_kompetensordboken
  n_notes_pappa_yaml_frolistan_kan_aterskapas_ur_masterdatabasen --> n_notes_pappa_dubblettknapp_mansklig_granskning_och_bestandig_sammanslagning
  n_notes_pappas_dev_miljo_kor_mot_neon_i_stallet_for_in_kluster_postgres --> n_notes_ci_saknade_inloggning_mot_ghcr_alla_image_pushar_misslyckades
  n_notes_pappas_dev_miljo_kor_mot_neon_i_stallet_for_in_kluster_postgres --> n_notes_nuxt_4_webbapp_for_pappa_laser_personerna_ur_neon
  n_notes_pappas_dev_miljo_kor_mot_neon_i_stallet_for_in_kluster_postgres --> n_notes_pappa_kor_nu_pa_neon_neon_ts_env_local_och_poolad_direkt_wiring
  n_notes_pappas_dev_miljo_kor_mot_neon_i_stallet_for_in_kluster_postgres --> n_notes_plan_dev_kor_mot_neon_istallet_for_lokal_postgres
  n_notes_pappas_dev_miljo_kor_mot_neon_i_stallet_for_in_kluster_postgres --> n_notes_plan_repots_ci_och_prod_pipeline_var_trasig_fixade_sokvagar_och_bygg
  n_notes_pappas_dev_miljo_kor_mot_neon_i_stallet_for_in_kluster_postgres --> n_notes_repo_template_ci_gitops_flodet_fixat_bygger_3_images_ghcr_tagging_pr_overlays_per_pr
  n_notes_pappas_dev_miljo_kor_mot_neon_i_stallet_for_in_kluster_postgres --> n_notes_stk_dev_miljon_kor_mot_neon_istallet_for_cnpg
  n_notes_pappas_dev_miljo_kor_mot_neon_i_stallet_for_in_kluster_postgres --> n_notes_tilt_filbevakning_kan_do_starta_om_tilt_up
  n_notes_pappas_dev_miljo_kor_mot_neon_i_stallet_for_in_kluster_postgres --> n_notes_transienta_502_fra_go_api_t_nar_poddar_startar_om_i_dev_klustret
  n_notes_pappas_grundare_ar_michael_brundin_inte_git_anvandaren_simon_brundin --> n_notes_redigerare_och_agarskap_for_relationer_i_pappa
  n_notes_parallella_agent_sessioner_i_samma_repo_och_databas --> n_notes_kalenderns_nivavaxling_animeras_crossfade_pop_in_och_hysteresis
  n_notes_parallella_agent_sessioner_i_samma_repo_och_databas --> n_notes_vertikal_kalender_med_kontinuerlig_zoom
  n_notes_personer_kan_laggas_in_manuellt_i_pappa --> n_notes_pappa_dubblettknapp_mansklig_granskning_och_bestandig_sammanslagning
  n_notes_personer_kan_laggas_in_manuellt_i_pappa --> n_notes_relationsoversikt_i_pappa_var_ser_jag_mina_egna_relationer
  n_notes_photocraft_agentstyrning_via_control_protokoll_control_7878 --> n_notes_photocraft_kamerans_look_varning_vid_raw_oppning_och_hur_den_hanteras
  n_notes_photocraft_kamerans_look_varning_vid_raw_oppning_och_hur_den_hanteras --> n_notes_lightcraft_photocraft_installerade_pa_omarchy_workstation_utan_sudo
  n_notes_pi_durable_1_4_0_api_avvikelser_fran_blogginlagget_och_tva_gotchas --> n_notes_beslut_pi_durable_satts_pa_hyllan_for_tillfallet
  n_notes_pi_durable_1_4_0_api_avvikelser_fran_blogginlagget_och_tva_gotchas --> n_notes_pi_durable_demo_paketet_earendil_works_pi_durable_fungerar
  n_notes_pi_durable_1_4_0_api_avvikelser_fran_blogginlagget_och_tva_gotchas --> n_notes_pi_durable_fungerar_med_minimax_coding_plan_anthropic_messages_api
  n_notes_pi_durable_demo_paketet_earendil_works_pi_durable_fungerar --> n_notes_beslut_pi_durable_satts_pa_hyllan_for_tillfallet
  n_notes_pi_durable_demo_paketet_earendil_works_pi_durable_fungerar --> n_notes_pi_durable_1_4_0_api_avvikelser_fran_blogginlagget_och_tva_gotchas
  n_notes_pi_durable_demo_paketet_earendil_works_pi_durable_fungerar --> n_notes_pi_durable_fungerar_med_minimax_coding_plan_anthropic_messages_api
  n_notes_pi_durable_fungerar_med_minimax_coding_plan_anthropic_messages_api --> n_notes_beslut_pi_durable_satts_pa_hyllan_for_tillfallet
  n_notes_pi_durable_fungerar_med_minimax_coding_plan_anthropic_messages_api --> n_notes_pi_durable_1_4_0_api_avvikelser_fran_blogginlagget_och_tva_gotchas
  n_notes_pi_durable_fungerar_med_minimax_coding_plan_anthropic_messages_api --> n_notes_pi_durable_demo_paketet_earendil_works_pi_durable_fungerar
  n_notes_picokvm_luckfox_startar_inte_i_labbet_lite_saknar_emmc_ingen_poe --> n_notes_u7_lite_94_2a_6f_7a_4b_33_adopterar_inte_saknar_management_ip_och_nattverkssegment_pa_switchen
  n_notes_plan_appen_morkt_designsystem_med_semantiska_farg_token --> n_notes_mailvyn_mail_omstegad_till_tailwind_tema
  n_notes_plan_repots_ci_och_prod_pipeline_var_trasig_fixade_sokvagar_och_bygg --> n_notes_authflytten_i_plan_sessionen_ligger_nu_i_go_api_t_inte_i_nuxt
  n_notes_platser_pa_mal_och_kalenderhandelser_openspec_change_add_places --> n_notes_malgraph_navigering_mellan_malsidor_gar_fran_sekunder_till_millisekunder
  n_notes_popcorn_time_installerad_som_appimage_i_local --> n_notes_popcorn_time_remote_movies_api_failed_to_respond_orsak_och_fix
  n_notes_popcorn_time_remote_movies_api_failed_to_respond_orsak_och_fix --> n_notes_popcorn_time_svenskt_ljud_finns_inte_i_katalogen_anvand_torrentsamling_torrent_collection
  n_notes_popcorn_time_svenskt_ljud_finns_inte_i_katalogen_anvand_torrentsamling_torrent_collection --> n_notes_qbittorrent_ar_standardhanterare_for_torrent_och_magnet_lankar
  n_notes_prioriteringsordningen_pa_now_sparades_inte_bulk_update_och_spar_ko --> n_notes_sortering_av_grundmal_pa_goals_med_piltangenterna
  n_notes_priorityfilter_ska_behandla_saknat_started_som_ej_paborjat --> n_notes_mal_kan_ha_due_datum_deadline
  n_notes_pxe_aterstallning_nodspecifika_matchbox_profiler_och_maintenance_mode_flode --> n_entities_pxe_server
  n_notes_pxe_aterstallning_nodspecifika_matchbox_profiler_och_maintenance_mode_flode --> n_notes_recovery_av_en_rock_worker_kraver_fysisk_rattning_esp_filen_pa_emmc_fangar_booten
  n_notes_pxe_boot_script_dhcp_kraver_mkimage_uimage_inte_ren_hush_text --> n_entities_pxe_server
  n_notes_pxe_boot_script_dhcp_kraver_mkimage_uimage_inte_ren_hush_text --> n_notes_u_boot_for_rock_4c_built_from_mainline_with_pxe_first
  n_notes_pxe_boot_script_dhcp_kraver_mkimage_uimage_inte_ren_hush_text --> n_notes_worker_1_kortet_flashat_med_ny_mainline_build_tf_a_fran_repot
  n_notes_pxe_boot_script_dhcp_kraver_mkimage_uimage_inte_ren_hush_text --> n_notes_worker_1_rock_4c_tyst_mainline_u_boot_byggd_men_boarden_laser_inte_kortet
  n_notes_pxe_bootloop_initramfs_startar_om_for_att_boota_installerad_talos_pa_emmc --> n_notes_rescue_wipe_pa_rock_maste_tomma_hela_systemdisken_wipe_system_esp_filen_ar_det_som_fangar_booten
  n_notes_pxe_bootloop_initramfs_startar_om_for_att_boota_installerad_talos_pa_emmc --> n_notes_talos_v1_14_utan_unattendedinstallconfig_oinstallerad_metal_nod_bootar_om_i_en_loop
  n_notes_pxe_forst_bootloader_for_rock_4c_byggd_som_patchad_u_boot_image --> n_entities_pxe_server
  n_notes_pxe_forst_bootloader_for_rock_4c_byggd_som_patchad_u_boot_image --> n_notes_rock_4c_och_4se_har_ingen_spi_flash_i_fabrik_u_boot_ligger_pa_microsd_emmc
  n_notes_pxe_forst_bootloader_for_rock_4c_byggd_som_patchad_u_boot_image --> n_notes_u_boot_kors_en_gang_boot_targets_styr_steg_2_inte_steg_1
  n_notes_pxe_forst_bootloader_for_rock_4c_byggd_som_patchad_u_boot_image --> n_notes_worker_1_rock_4c_nar_pxe_via_u_boot_boot_targets_men_pxe_ligger_sist
  n_notes_pxe_genererade_nodconfigs_var_stale_efter_talos_v1_14_uppgraderingen --> n_entities_pxe_server
  n_notes_pxe_genererade_nodconfigs_var_stale_efter_talos_v1_14_uppgraderingen --> n_entities_talos_linux
  n_notes_pxe_genererade_nodconfigs_var_stale_efter_talos_v1_14_uppgraderingen --> n_notes_rock_4c_4se_utan_batteribackup_rtc_star_pa_2017_tls_mot_factory_talos_dev_misslyckas
  n_notes_pxe_genererade_nodconfigs_var_stale_efter_talos_v1_14_uppgraderingen --> n_notes_talos_v1_14_machineconfig_specen_ar_en_quoted_string_och_flera_dokument
  n_notes_pxe_kedjan_dhcp_boot_pekar_pa_textskript_matchbox_efi_ar_uefi_vagen --> n_entities_pxe_server
  n_notes_pxe_kedjan_dhcp_boot_pekar_pa_textskript_matchbox_efi_ar_uefi_vagen --> n_notes_pxe_forst_bootloader_for_rock_4c_byggd_som_patchad_u_boot_image
  n_notes_pxe_kedjan_dhcp_boot_pekar_pa_textskript_matchbox_efi_ar_uefi_vagen --> n_notes_pxe_varden_korde_foraldrad_matchbox_nodconfiger_saknades_404
  n_notes_pxe_kedjan_dhcp_boot_pekar_pa_textskript_matchbox_efi_ar_uefi_vagen --> n_notes_u_boot_kors_en_gang_boot_targets_styr_steg_2_inte_steg_1
  n_notes_pxe_kedjan_for_worker_1_verifierad_serverside_pxe_metoden_failar_snabbt_utan_pxelinux_cfg --> n_entities_pxe_server
  n_notes_pxe_kedjan_for_worker_1_verifierad_serverside_pxe_metoden_failar_snabbt_utan_pxelinux_cfg --> n_notes_u_boot_for_rock_4c_built_from_mainline_with_pxe_first
  n_notes_pxe_kedjan_for_worker_1_verifierad_serverside_pxe_metoden_failar_snabbt_utan_pxelinux_cfg --> n_notes_worker_1_kortet_flashat_med_ny_mainline_build_tf_a_fran_repot
  n_notes_pxe_kedjan_for_worker_1_verifierad_serverside_pxe_metoden_failar_snabbt_utan_pxelinux_cfg --> n_notes_worker_1_rock_4c_tyst_mainline_u_boot_byggd_men_boarden_laser_inte_kortet
  n_notes_pxe_mainline_u_boots_bootefi_kraver_addr_size_for_image_och_initrd --> n_entities_pxe_server
  n_notes_pxe_mainline_u_boots_bootefi_kraver_addr_size_for_image_och_initrd --> n_notes_worker_1_kortet_flashat_med_ny_mainline_build_tf_a_fran_repot
  n_notes_pxe_mainline_u_boots_bootefi_kraver_addr_size_for_image_och_initrd --> n_notes_worker_1_rock_4c_nar_pxe_via_u_boot_boot_targets_men_pxe_ligger_sist
  n_notes_pxe_mainline_u_boots_bootefi_kraver_addr_size_for_image_och_initrd --> n_notes_worker_1_rock_4c_tyst_mainline_u_boot_byggd_men_boarden_laser_inte_kortet
  n_notes_pxe_status_fungerar_i_produktionsklustret_tftp_svarar_inte_fran_pod_natet --> n_entities_pxe_server
  n_notes_pxe_status_fungerar_i_produktionsklustret_tftp_svarar_inte_fran_pod_natet --> n_notes_pxe_status_pa_dashboard_startsidan_tftp_gar_inte_att_kontrollera_fran_pod_natet
  n_notes_pxe_status_pa_dashboard_startsidan_tftp_gar_inte_att_kontrollera_fran_pod_natet --> n_entities_pxe_server
  n_notes_pxe_status_pa_dashboard_startsidan_tftp_gar_inte_att_kontrollera_fran_pod_natet --> n_notes_pxe_status_fungerar_i_produktionsklustret_tftp_svarar_inte_fran_pod_natet
  n_notes_pxe_varden_korde_foraldrad_matchbox_nodconfiger_saknades_404 --> n_entities_pxe_server
  n_notes_pxe_varden_korde_foraldrad_matchbox_nodconfiger_saknades_404 --> n_notes_pxe_kedjan_dhcp_boot_pekar_pa_textskript_matchbox_efi_ar_uefi_vagen
  n_notes_qbittorrent_ar_standardhanterare_for_torrent_och_magnet_lankar --> n_notes_qbittorrent_installerad_pa_workstation_omarchy_arch
  n_notes_qbittorrent_installerad_pa_workstation_omarchy_arch --> n_notes_popcorn_time_installerad_som_appimage_i_local
  n_notes_radxa_bradena_behover_bootmedium_men_talos_och_bootloader_maste_ligga_pa_skilda_medium --> n_notes_en_talos_efi_partition_pa_kortet_blockerar_pxe_booten_bootflow_loaded_fore_natverket
  n_notes_radxa_bradena_behover_bootmedium_men_talos_och_bootloader_maste_ligga_pa_skilda_medium --> n_notes_gotcha_pxe_forst_gor_noden_till_en_pxe_nod_maintenance_mode
  n_notes_radxa_bradena_behover_bootmedium_men_talos_och_bootloader_maste_ligga_pa_skilda_medium --> n_notes_pa_rock_plattorna_finns_ingen_bootloader_i_hardvaran
  n_notes_radxa_bradena_behover_bootmedium_men_talos_och_bootloader_maste_ligga_pa_skilda_medium --> n_notes_rk3399_s_bootrom_soker_spi_emmc_sd_usb_hardkodat
  n_notes_radxa_bradena_behover_bootmedium_men_talos_och_bootloader_maste_ligga_pa_skilda_medium --> n_notes_talos_installen_nollstaller_de_forsta_11_mb_pa_microsd_loadern_maste_aterstallas
  n_notes_rantebetalningsmejl_hamnar_i_inbox_deleted_messages_och_arkivet_matchades_aldrig --> n_notes_kameo_ranteutbetalning_redan_utbetald_samlingsuttag_pa_47_29_kr
  n_notes_redigerare_och_agarskap_for_relationer_i_pappa --> n_notes_detaljvy_for_kandidat_i_pappa_gemensamma_vagar_mellan_kandidater
  n_notes_redigerare_och_agarskap_for_relationer_i_pappa --> n_notes_min_kontaktlista_jag_som_nod_i_grafen_pappa
  n_notes_redigerare_och_agarskap_for_relationer_i_pappa --> n_notes_pappas_grundare_ar_michael_brundin_inte_git_anvandaren_simon_brundin
  n_notes_redigerare_och_agarskap_for_relationer_i_pappa --> n_notes_relationsoversikt_i_pappa_var_ser_jag_mina_egna_relationer
  n_notes_rekrytering_av_it_personal_i_sundsvall_marknadsbild_och_frolista --> n_notes_it_arbetsplatser_i_sundsvall_3_469_enligt_scb_yrkesregistret
  n_notes_rekrytering_av_it_personal_i_sundsvall_marknadsbild_och_frolista --> n_notes_masterdatabas_i_postgres_schema_import_och_drift_for_it_rekryteringen
  n_notes_rekrytering_av_it_personal_i_sundsvall_marknadsbild_och_frolista --> n_notes_pappa_anstallningshistorik_i_cv_format_fix_av_kompetensordboken
  n_notes_rekrytering_av_it_personal_i_sundsvall_marknadsbild_och_frolista --> n_notes_utbyggnad_av_pappas_personlista_bolagsregister_konferensbios_gav_140_namngivna_personer
  n_notes_relationer_kan_nu_laggas_till_fran_oversikten_dialogen_har_par_lage --> n_notes_relationsoversikt_i_pappa_var_ser_jag_mina_egna_relationer
  n_notes_relationer_mellan_kandidater_i_pappa_symmetriska_par_med_styrka_1_10 --> n_notes_detaljvy_for_kandidat_i_pappa_gemensamma_vagar_mellan_kandidater
  n_notes_relationer_mellan_kandidater_i_pappa_symmetriska_par_med_styrka_1_10 --> n_notes_pappa_dubblettknapp_mansklig_granskning_och_bestandig_sammanslagning
  n_notes_relationer_mellan_kandidater_i_pappa_symmetriska_par_med_styrka_1_10 --> n_notes_rekryteringspipeline_board_per_person_sourcing_som_attribut
  n_notes_relationsgrafen_anvander_fast_michael_id_utan_person_kind --> n_notes_michael_brundin_blir_en_personnod_i_relationsgrafen
  n_notes_relationsoversikt_i_pappa_var_ser_jag_mina_egna_relationer --> n_notes_redigerare_och_agarskap_for_relationer_i_pappa
  n_notes_relationsoversikt_i_pappa_var_ser_jag_mina_egna_relationer --> n_notes_relationer_kan_nu_laggas_till_fran_oversikten_dialogen_har_par_lage
  n_notes_rk3399_s_bootrom_soker_spi_emmc_sd_usb_hardkodat --> n_entities_pxe_server
  n_notes_rk3399_s_bootrom_soker_spi_emmc_sd_usb_hardkodat --> n_notes_pa_rock_plattorna_finns_ingen_bootloader_i_hardvaran
  n_notes_rk3399_s_bootrom_soker_spi_emmc_sd_usb_hardkodat --> n_notes_u_boot_2026_01_rock_4c_saknar_bootfile_boot_script_dhcp_natverksboot_dor_direkt
  n_notes_rk3399_s_bootrom_soker_spi_emmc_sd_usb_hardkodat --> n_notes_u_boot_kors_en_gang_boot_targets_styr_steg_2_inte_steg_1
  n_notes_rock_4c_och_4se_har_ingen_spi_flash_i_fabrik_u_boot_ligger_pa_microsd_emmc --> n_entities_pxe_server
  n_notes_rock_4c_och_4se_har_ingen_spi_flash_i_fabrik_u_boot_ligger_pa_microsd_emmc --> n_notes_u_boot_2026_01_rock_4c_saknar_bootfile_boot_script_dhcp_natverksboot_dor_direkt
  n_notes_rock_4c_och_4se_har_ingen_spi_flash_i_fabrik_u_boot_ligger_pa_microsd_emmc --> n_notes_u_boot_for_rock_4c_built_from_mainline_with_pxe_first
  n_notes_rock_4c_och_4se_har_ingen_spi_flash_i_fabrik_u_boot_ligger_pa_microsd_emmc --> n_notes_worker_1_rock_4c_nar_pxe_via_u_boot_boot_targets_men_pxe_ligger_sist
  n_notes_rock_pi_4_rk3399_defconfig_bygger_en_rock_pi_4a_for_4c_ska_rock_4c_plus_rk3399_defconfig_anvandas --> n_entities_pxe_server
  n_notes_rock_pi_4_rk3399_defconfig_bygger_en_rock_pi_4a_for_4c_ska_rock_4c_plus_rk3399_defconfig_anvandas --> n_notes_pxe_kedjan_for_worker_1_verifierad_serverside_pxe_metoden_failar_snabbt_utan_pxelinux_cfg
  n_notes_rock_pi_4_rk3399_defconfig_bygger_en_rock_pi_4a_for_4c_ska_rock_4c_plus_rk3399_defconfig_anvandas --> n_notes_u_boot_for_rock_4c_built_from_mainline_with_pxe_first
  n_notes_rock_pi_4_rk3399_defconfig_bygger_en_rock_pi_4a_for_4c_ska_rock_4c_plus_rk3399_defconfig_anvandas --> n_notes_worker_1_kortet_flashat_med_ny_mainline_build_tf_a_fran_repot
  n_notes_root_dir_funktionen_stoppar_no_project_root_found_meddelandet --> n_notes_rust_analyzer_standalone_lage_kraschar_pa_leetcode_nvim_filer
  n_notes_rust_analyzer_standalone_lage_kraschar_pa_leetcode_nvim_filer --> n_notes_rust_analyzer_tysta_health_status_notisen_istallet_for_att_stanga_av_standalone
  n_notes_rust_analyzer_tysta_health_status_notisen_istallet_for_att_stanga_av_standalone --> n_notes_sa_dampas_bada_rust_analyzer_popupen_no_project_root_found
  n_notes_rust_learning_repot_lades_in_som_submodule_mot_egen_fork --> n_notes_rust_repo_learn_to_code_with_rust_som_submodule
  n_notes_rust_repo_learn_to_code_with_rust_som_submodule --> n_notes_rust_utvecklingsmiljo_pa_workstationen
  n_notes_rust_utvecklingsmiljo_pa_workstationen --> n_notes_root_dir_funktionen_stoppar_no_project_root_found_meddelandet
  n_notes_rust_utvecklingsmiljo_pa_workstationen --> n_notes_rust_learning_repot_lades_in_som_submodule_mot_egen_fork
  n_notes_session_livscykel --> n_notes_accuracy_fel_omstart
  n_notes_session_livscykel --> n_notes_arkitektur
  n_notes_session_livscykel --> n_notes_sprint_timer_saknas
  n_notes_session_livscykel --> n_notes_accuracy_fel_omstart
  n_notes_session_livscykel --> n_notes_arkitektur
  n_notes_simon_kvm_cli_kommando_for_att_byta_hdmi_kvm_port_fran_pxe_vard --> n_entities_hdmi_kvm_switch_16_port_labbet
  n_notes_simon_kvm_cli_kommando_for_att_byta_hdmi_kvm_port_fran_pxe_vard --> n_notes_kvm_portmappning_worker_n_port_n_den_deployade_configen_pa_pi_n_var_fel
  n_notes_simon_kvm_cli_kommando_for_att_byta_hdmi_kvm_port_fran_pxe_vard --> n_notes_kvm_styrning_fran_pxe_pi_n_16_port_hdmi_kvm
  n_notes_simon_pxe_flash_clean_code_omgang_2_fokuserade_funktioner --> n_entities_pxe_server
  n_notes_simon_pxe_flash_flashar_pxe_forst_u_boot_till_sd_kort_i_pxe_servern --> n_entities_pxe_server
  n_notes_simon_pxe_flash_flashar_pxe_forst_u_boot_till_sd_kort_i_pxe_servern --> n_notes_worker_1_kortet_flashat_med_ny_mainline_build_tf_a_fran_repot
  n_notes_simon_pxe_flash_flashar_pxe_forst_u_boot_till_sd_kort_i_pxe_servern --> n_notes_worker_1_rock_4c_nar_pxe_via_u_boot_boot_targets_men_pxe_ligger_sist
  n_notes_simon_pxe_flash_flashar_pxe_forst_u_boot_till_sd_kort_i_pxe_servern --> n_notes_worker_1_rock_4c_tyst_mainline_u_boot_byggd_men_boarden_laser_inte_kortet
  n_notes_skogsduvan_elementet_styrs_av_minsta_temperaturen_fail_open --> n_notes_skogsduvan_temperaturkaskad_i_tre_steg_styrd_av_minsta_temperaturen
  n_notes_skogsduvan_fyra_shelly_enheter_svarar_inte_pa_natet --> n_notes_skogsduvans_ha_temperaturlogik_ombyggd_fyra_doda_sensorer
  n_notes_skogsduvan_larm_nar_sensorer_shelly_enheter_faller_ur_online --> n_notes_skogsduvan_temperaturkaskad_i_tre_steg_styrd_av_minsta_temperaturen
  n_notes_skogsduvan_ssh_via_duckdns_fungerar_port_22_forwardad_2026_11_02 --> n_notes_twingate_anslutaren_for_10_10_10_0_24_var_nere_recovery_floden_kan_inte_koras_fran_arbetsstationen
  n_notes_skogsduvan_temperaturkaskad_i_tre_steg_styrd_av_minsta_temperaturen --> n_notes_skogsduvan_larm_nar_sensorer_shelly_enheter_faller_ur_online
  n_notes_skogsduvan_temperaturkaskad_i_tre_steg_styrd_av_minsta_temperaturen --> n_notes_skogsduvan_ssh_via_duckdns_fungerar_port_22_forwardad_2026_11_02
  n_notes_skogsduvan_temperaturkaskad_i_tre_steg_styrd_av_minsta_temperaturen --> n_notes_skogsduvans_ha_temperaturlogik_ombyggd_fyra_doda_sensorer
  n_notes_skogsduvan_temperaturkaskad_i_tre_steg_styrd_av_minsta_temperaturen --> n_notes_skogsduvan_elementet_styrs_av_minsta_temperaturen_fail_open
  n_notes_skogsduvan_temperaturkaskad_i_tre_steg_styrd_av_minsta_temperaturen --> n_notes_ha_automationer_behover_en_startv_g_for_att_aterstalla_laget_och_tre_syntaxfallor
  n_notes_skogsduvan_temperaturkaskad_i_tre_steg_styrd_av_minsta_temperaturen --> n_notes_skogsduvans_ha_utsatts_for_automatiserade_skannningar_fran_internet
  n_notes_skogsduvans_ha_temperaturlogik_ombyggd_fyra_doda_sensorer --> n_notes_skogsduvan_fyra_shelly_enheter_svarar_inte_pa_natet
  n_notes_skogsduvans_ha_temperaturlogik_ombyggd_fyra_doda_sensorer --> n_notes_skogsduvan_temperaturkaskad_i_tre_steg_styrd_av_minsta_temperaturen
  n_notes_skogsduvans_ha_utsatts_for_automatiserade_skannningar_fran_internet --> n_notes_skogsduvans_ha_temperaturlogik_ombyggd_fyra_doda_sensorer
  n_notes_sortering_av_grundmal_pa_goals_med_piltangenterna --> n_notes_prioriteringsordningen_pa_now_sparades_inte_bulk_update_och_spar_ko
  n_notes_sortering_av_grundmal_pa_goals_med_piltangenterna --> n_notes_sortering_av_undermal_med_piltangenterna_var_trasig_i_plan
  n_notes_sprint_timer_saknas --> n_notes_session_livscykel
  n_notes_sprint_timer_saknas --> n_notes_session_livscykel
  n_notes_standardkalender_i_plan --> n_notes_flera_kalendrar_med_egna_farger_i_plan
  n_notes_stk_matchi_synk_sjalvaterstallande_synk_och_var_database_url_satts --> n_notes_synk_hallbokningar_kommunens_actor_smartbook_matchi
  n_notes_stray_wiki_i_typetrainer_repot_ar_inte_kallan_for_wikin --> n_notes_kunskaps_wiki
  n_notes_synk_hallbokningar_kommunens_actor_smartbook_matchi --> n_notes_gron_markering_av_tider_som_ar_bokade_i_matchi_skrivlasande_jamforelse
  n_notes_synk_hallbokningar_kommunens_actor_smartbook_matchi --> n_notes_krypterade_secrets_i_klustret_ogiltiga_inloggningar_och_keycloak_brute_force
  n_notes_synk_hallbokningar_kommunens_actor_smartbook_matchi --> n_notes_matchi_bokning_slot_id_andras_nar_bokningen_skapas
  n_notes_synk_hallbokningar_kommunens_actor_smartbook_matchi --> n_notes_matchi_keycloak_session_code_maste_fangas_med_generell_teckenklass
  n_notes_synk_hallbokningar_kommunens_actor_smartbook_matchi --> n_notes_query_parametrar_i_actor_smartbook_endpointen_url_avkodas_till_mellanslag
  n_notes_synk_hallbokningar_kommunens_actor_smartbook_matchi --> n_notes_stk_matchi_synk_sjalvaterstallande_synk_och_var_database_url_satts
  n_notes_talos_v1_14_machineconfig_specen_ar_en_quoted_string_och_flera_dokument --> n_entities_talos_linux
  n_notes_talos_v1_14_machineconfig_specen_ar_en_quoted_string_och_flera_dokument --> n_notes_pxe_genererade_nodconfigs_var_stale_efter_talos_v1_14_uppgraderingen
  n_notes_talos_v1_14_machineconfig_specen_ar_en_quoted_string_och_flera_dokument --> n_notes_talos_hostname_auto_i_en_generisk_config_byter_nodens_namn_och_kastar_ut_den_ur_klustret
  n_notes_talos_v1_14_machineconfig_specen_ar_en_quoted_string_och_flera_dokument --> n_summaries_longhorn_disk_guide
  n_notes_tidsatgang_pa_mal_i_plan_estimated_minutes --> n_notes_plan_dev_kor_mot_neon_istallet_for_lokal_postgres
  n_notes_tidsatgang_pa_mal_i_plan_estimated_minutes --> n_notes_undo_stapel_i_plan_z_angrar_senaste_malhandling
  n_notes_tilt_filbevakning_kan_do_starta_om_tilt_up --> n_notes_repo_template_rust_api_kopplad_till_kubernetes_manifests_tilt
  n_notes_tre_satt_att_pxe_boota_rock_4c_och_varfor_dhcp_metoden_valdes --> n_entities_pxe_server
  n_notes_tre_satt_att_pxe_boota_rock_4c_och_varfor_dhcp_metoden_valdes --> n_notes_u_boot_for_rock_4c_built_from_mainline_with_pxe_first
  n_notes_tre_satt_att_pxe_boota_rock_4c_och_varfor_dhcp_metoden_valdes --> n_notes_worker_1_kortet_flashat_med_ny_mainline_build_tf_a_fran_repot
  n_notes_tre_satt_att_pxe_boota_rock_4c_och_varfor_dhcp_metoden_valdes --> n_notes_worker_1_rock_4c_tyst_mainline_u_boot_byggd_men_boarden_laser_inte_kortet
  n_notes_twingate_anslutaren_for_10_10_10_0_24_var_nere_recovery_floden_kan_inte_koras_fran_arbetsstationen --> n_notes_twingate_ressurserna_forsvann_pa_arbetsstationen_2026_11_04_hemnatet_10_1_1_1_och_datacenter_10_10_10_0_24_oppnarbara
  n_notes_u_boot_2026_01_rock_4c_saknar_bootfile_boot_script_dhcp_natverksboot_dor_direkt --> n_entities_pxe_server
  n_notes_u_boot_2026_01_rock_4c_saknar_bootfile_boot_script_dhcp_natverksboot_dor_direkt --> n_notes_pa_rock_plattorna_finns_ingen_bootloader_i_hardvaran
  n_notes_u_boot_2026_01_rock_4c_saknar_bootfile_boot_script_dhcp_natverksboot_dor_direkt --> n_notes_rk3399_s_bootrom_soker_spi_emmc_sd_usb_hardkodat
  n_notes_u_boot_2026_01_rock_4c_saknar_bootfile_boot_script_dhcp_natverksboot_dor_direkt --> n_notes_u_boot_kors_en_gang_boot_targets_styr_steg_2_inte_steg_1
  n_notes_u_boot_for_rock_4c_built_from_mainline_with_pxe_first --> n_entities_pxe_server
  n_notes_u_boot_for_rock_4c_built_from_mainline_with_pxe_first --> n_notes_pxe_kedjan_for_worker_1_verifierad_serverside_pxe_metoden_failar_snabbt_utan_pxelinux_cfg
  n_notes_u_boot_for_rock_4c_built_from_mainline_with_pxe_first --> n_notes_rock_4c_4se_talos_kan_aldrig_installeras_till_disk_u_boot_saknar_efi_setvariablert
  n_notes_u_boot_for_rock_4c_built_from_mainline_with_pxe_first --> n_notes_worker_1_kortet_flashat_med_ny_mainline_build_tf_a_fran_repot
  n_notes_u_boot_for_rock_4c_built_from_mainline_with_pxe_first --> n_notes_worker_1_rock_4c_tyst_mainline_u_boot_byggd_men_boarden_laser_inte_kortet
  n_notes_u_boot_hamtar_alltid_boot_script_dhcp_per_nod_config_via_ipaddr --> n_entities_pxe_server
  n_notes_u_boot_hamtar_alltid_boot_script_dhcp_per_nod_config_via_ipaddr --> n_notes_worker_1_kortet_flashat_med_ny_mainline_build_tf_a_fran_repot
  n_notes_u_boot_hamtar_alltid_boot_script_dhcp_per_nod_config_via_ipaddr --> n_notes_worker_1_rock_4c_nar_pxe_via_u_boot_boot_targets_men_pxe_ligger_sist
  n_notes_u_boot_hamtar_alltid_boot_script_dhcp_per_nod_config_via_ipaddr --> n_notes_worker_1_rock_4c_tyst_mainline_u_boot_byggd_men_boarden_laser_inte_kortet
  n_notes_u_boot_kors_en_gang_boot_targets_styr_steg_2_inte_steg_1 --> n_entities_pxe_server
  n_notes_u_boot_kors_en_gang_boot_targets_styr_steg_2_inte_steg_1 --> n_notes_pxe_forst_bootloader_for_rock_4c_byggd_som_patchad_u_boot_image
  n_notes_u_boot_kors_en_gang_boot_targets_styr_steg_2_inte_steg_1 --> n_notes_rk3399_s_bootrom_soker_spi_emmc_sd_usb_hardkodat
  n_notes_u_boot_kors_en_gang_boot_targets_styr_steg_2_inte_steg_1 --> n_notes_u_boot_2026_01_rock_4c_saknar_bootfile_boot_script_dhcp_natverksboot_dor_direkt
  n_notes_undo_stapel_i_plan_z_angrar_senaste_malhandling --> n_notes_tidsatgang_pa_mal_i_plan_estimated_minutes
  n_notes_unifi_gatewayen_har_ingen_publik_ipv6_cgnat_via_starlink --> n_notes_ipv6_ingress_till_envoy_designlage_2026_11_15
  n_notes_usb_disken_acer_wd_320gb_status_och_longhorn_bedomning --> n_concepts_longhorn
  n_notes_usb_disken_acer_wd_320gb_status_och_longhorn_bedomning --> n_entities_talos_linux
  n_notes_utbyggnad_av_pappas_personlista_bolagsregister_konferensbios_gav_140_namngivna_personer --> n_notes_pappa_anstallningshistorik_i_cv_format_fix_av_kompetensordboken
  n_notes_utbyggnad_av_pappas_personlista_bolagsregister_konferensbios_gav_140_namngivna_personer --> n_notes_rekrytering_av_it_personal_i_sundsvall_marknadsbild_och_frolista
  n_notes_uttunnade_serier_vags_sa_att_antalen_blir_exakta --> n_notes_aterkommande_serier_doljs_inte_langre_beroende_pa_zoomniva
  n_notes_utvecklingsflode --> n_notes_arbetsstationen_ar_omarchy_arch_apt_get_finns_inte_anvand_pacman_yay
  n_notes_utvecklingsflode --> n_notes_arkitektur
  n_notes_utvecklingsflode --> n_notes_konventioner
  n_notes_utvecklingsflode --> n_summaries_llm_wiki_karpathy
  n_notes_utvecklingsflode --> n_notes_kunskaps_wiki
  n_notes_utvecklingsflode --> n_notes_konventioner
  n_notes_valfri_fargkod_pa_kalendrar --> n_notes_flera_kalendrar_med_egna_farger_i_plan
  n_notes_verifierade_policy_hal_i_tool_lagret_kommandon_ar_blocklist_inte_allowlist --> n_notes_artikeln_om_agentic_software_factory_vad_som_ar_relevant_for_agentsystemet
  n_notes_vertikal_kalender_med_ikonstod_i_plan --> n_notes_flera_kalendrar_med_egna_farger_i_plan
  n_notes_vertikal_kalender_med_ikonstod_i_plan --> n_notes_ikonvaljaren_i_plan_api_icons_routen_var_borttagen_och_ikonprovider_none_gjorde_att_inga_ikoner_renderades
  n_notes_vertikal_kalender_med_ikonstod_i_plan --> n_notes_kalenderns_detaljnivaer_fem_nivaer_och_lane_packning
  n_notes_vertikal_kalender_med_ikonstod_i_plan --> n_notes_kalenderns_zoom_styrs_av_vertikal_scrollning
  n_notes_vertikal_kalender_med_ikonstod_i_plan --> n_notes_nya_mal_far_ikon_fran_titeln
  n_notes_vertikal_kalender_med_ikonstod_i_plan --> n_notes_plan_api_anvander_scalar_for_openapi
  n_notes_vertikal_kalender_med_ikonstod_i_plan --> n_notes_repo_template_ny_rust_api_src_rust_med_axum_utoipa_scalar
  n_notes_vertikal_kalender_med_ikonstod_i_plan --> n_notes_vertikal_kalender_med_kontinuerlig_zoom
  n_notes_vertikal_kalender_med_kontinuerlig_zoom --> n_notes_kalendern_hamtar_alla_handelser_en_gang_och_expanderar_serier_lokalt
  n_notes_vertikal_kalender_med_kontinuerlig_zoom --> n_notes_kalenderns_nivavaxling_animeras_crossfade_pop_in_och_hysteresis
  n_notes_vertikal_kalender_med_kontinuerlig_zoom --> n_notes_parallella_agent_sessioner_i_samma_repo_och_databas
  n_notes_wordlist_kallor --> n_notes_bigram_algoritmen
  n_notes_wordlist_kallor --> n_notes_utvecklingsflode
  n_notes_wordlist_kallor --> n_notes_bigram_algoritmen
  n_notes_worker_1_kortet_flashat_med_ny_mainline_build_tf_a_fran_repot --> n_entities_pxe_server
  n_notes_worker_1_kortet_flashat_med_ny_mainline_build_tf_a_fran_repot --> n_notes_pxe_kedjan_for_worker_1_verifierad_serverside_pxe_metoden_failar_snabbt_utan_pxelinux_cfg
  n_notes_worker_1_kortet_flashat_med_ny_mainline_build_tf_a_fran_repot --> n_notes_u_boot_for_rock_4c_built_from_mainline_with_pxe_first
  n_notes_worker_1_kortet_flashat_med_ny_mainline_build_tf_a_fran_repot --> n_notes_worker_1_rock_4c_tyst_mainline_u_boot_byggd_men_boarden_laser_inte_kortet
  n_notes_worker_1_och_worker_2_ar_doda_pa_link_niva_pxe_kedjan_ar_inte_blockerande --> n_entities_pxe_server
  n_notes_worker_1_och_worker_2_ar_doda_pa_link_niva_pxe_kedjan_ar_inte_blockerande --> n_notes_worker_1_rock_4c_tyst_mainline_u_boot_byggd_men_boarden_laser_inte_kortet
  n_notes_worker_1_och_worker_2_ar_doda_pa_link_niva_pxe_kedjan_ar_inte_blockerande --> n_notes_worker_1_svarar_inte_i_natet_felsokningsstatus
  n_notes_worker_1_och_worker_2_ar_doda_pa_link_niva_pxe_kedjan_ar_inte_blockerande --> n_notes_worker_2_utan_strom_port_22_up_false_poe_power_0
  n_notes_worker_1_rock_4c_nar_pxe_via_u_boot_boot_targets_men_pxe_ligger_sist --> n_entities_pxe_server
  n_notes_worker_1_rock_4c_nar_pxe_via_u_boot_boot_targets_men_pxe_ligger_sist --> n_notes_pxe_forst_bootloader_for_rock_4c_byggd_som_patchad_u_boot_image
  n_notes_worker_1_rock_4c_nar_pxe_via_u_boot_boot_targets_men_pxe_ligger_sist --> n_notes_rock_4c_och_4se_har_ingen_spi_flash_i_fabrik_u_boot_ligger_pa_microsd_emmc
  n_notes_worker_1_rock_4c_nar_pxe_via_u_boot_boot_targets_men_pxe_ligger_sist --> n_notes_worker_1_rock_4c_tyst_mainline_u_boot_byggd_men_boarden_laser_inte_kortet
  n_notes_worker_1_rock_4c_tyst_mainline_u_boot_byggd_men_boarden_laser_inte_kortet --> n_entities_pxe_server
  n_notes_worker_1_rock_4c_tyst_mainline_u_boot_byggd_men_boarden_laser_inte_kortet --> n_notes_u_boot_for_rock_4c_built_from_mainline_with_pxe_first
  n_notes_worker_1_rock_4c_tyst_mainline_u_boot_byggd_men_boarden_laser_inte_kortet --> n_notes_worker_1_kortet_flashat_med_ny_mainline_build_tf_a_fran_repot
  n_notes_worker_1_rock_4c_tyst_mainline_u_boot_byggd_men_boarden_laser_inte_kortet --> n_notes_worker_1_och_worker_2_ar_doda_pa_link_niva_pxe_kedjan_ar_inte_blockerande
  n_notes_worker_1_startar_om_med_strom_men_skickar_noll_paket_fysisk_omstart_bevisar_samma_fel --> n_entities_pxe_server
  n_notes_worker_1_startar_om_med_strom_men_skickar_noll_paket_fysisk_omstart_bevisar_samma_fel --> n_notes_worker_1_och_worker_2_ar_doda_pa_link_niva_pxe_kedjan_ar_inte_blockerande
  n_notes_worker_1_startar_om_med_strom_men_skickar_noll_paket_fysisk_omstart_bevisar_samma_fel --> n_notes_worker_1_svarar_inte_ens_med_en_full_poe_cykel_noll_paket_i_bade_riktningar
  n_notes_worker_1_svarar_inte_ens_med_en_full_poe_cykel_noll_paket_i_bade_riktningar --> n_entities_pxe_server
  n_notes_worker_1_svarar_inte_ens_med_en_full_poe_cykel_noll_paket_i_bade_riktningar --> n_notes_worker_1_och_worker_2_ar_doda_pa_link_niva_pxe_kedjan_ar_inte_blockerande
  n_notes_worker_1_svarar_inte_ens_med_en_full_poe_cykel_noll_paket_i_bade_riktningar --> n_notes_worker_7_s_kubeletcertifikat_signerades_av_en_gammal_kubernetes_ca_som_finns_kvar_i_git_historiken
  n_notes_worker_1_svarar_inte_i_natet_felsokningsstatus --> n_entities_pxe_server
  n_notes_worker_1_svarar_inte_i_natet_felsokningsstatus --> n_notes_worker_1_kortet_flashat_med_ny_mainline_build_tf_a_fran_repot
  n_notes_worker_1_svarar_inte_i_natet_felsokningsstatus --> n_notes_worker_1_och_worker_2_ar_doda_pa_link_niva_pxe_kedjan_ar_inte_blockerande
  n_notes_worker_1_svarar_inte_i_natet_felsokningsstatus --> n_notes_worker_1_rock_4c_tyst_mainline_u_boot_byggd_men_boarden_laser_inte_kortet
  n_notes_worker_2_utan_strom_port_22_up_false_poe_power_0 --> n_entities_pxe_server
  n_notes_worker_2_utan_strom_port_22_up_false_poe_power_0 --> n_notes_worker_1_och_worker_2_ar_doda_pa_link_niva_pxe_kedjan_ar_inte_blockerande
  n_notes_worker_7_s_kubeletcertifikat_signerades_av_en_gammal_kubernetes_ca_som_finns_kvar_i_git_historiken --> n_notes_cluster1_stale_kubeconfig_ca_och_noll_workers_diagnos_2026_11_11
  n_notes_worker_7_s_kubeletcertifikat_signerades_av_en_gammal_kubernetes_ca_som_finns_kvar_i_git_historiken --> n_notes_worker_8_korrupt_xfs_pa_ephemeral_partitionen_gjorde_att_kubelet_aldrig_startade_fix_med_mkfs_xfs
  n_notes_worker_8_korrupt_xfs_pa_ephemeral_partitionen_gjorde_att_kubelet_aldrig_startade_fix_med_mkfs_xfs --> n_notes_cluster1_stale_kubeconfig_ca_och_noll_workers_diagnos_2026_11_11
  n_notes_worker_8_korrupt_xfs_pa_ephemeral_partitionen_gjorde_att_kubelet_aldrig_startade_fix_med_mkfs_xfs --> n_notes_worker_7_s_kubeletcertifikat_signerades_av_en_gammal_kubernetes_ca_som_finns_kvar_i_git_historiken
  n_notes_workers_bootar_inte_via_pxe_trots_fungerande_pxe_server_unifi_poe_styrning_via_rest_device --> n_entities_pxe_server
  n_notes_workers_bootar_inte_via_pxe_trots_fungerande_pxe_server_unifi_poe_styrning_via_rest_device --> n_notes_worker_1_startar_om_med_strom_men_skickar_noll_paket_fysisk_omstart_bevisar_samma_fel
  n_summaries_16x1_hdmi_kvm_rs232_lan_protokoll --> n_entities_hdmi_kvm_switch_16_port_labbet
  n_summaries_16x1_hdmi_kvm_rs232_lan_protokoll --> n_notes_kvm_styrning_fran_pxe_pi_n_16_port_hdmi_kvm
  n_summaries_flux_components --> n_concepts_flux_operator
  n_summaries_flux_components --> n_entities_flux_gitops
  n_summaries_flux_components --> n_entities_kubernetes
  n_summaries_infrastructure_flux --> n_concepts_flux_operator
  n_summaries_infrastructure_flux --> n_concepts_gitops
  n_summaries_kalenderns_detaljnivaer_och_lane_packning --> n_notes_kalenderns_detaljnivaer_fem_nivaer_och_lane_packning
  n_summaries_kalenderns_detaljnivaer_och_lane_packning --> n_notes_vertikal_kalender_med_ikonstod_i_plan
  n_summaries_kalenderns_detaljnivaer_och_lane_packning --> n_notes_vertikal_kalender_med_kontinuerlig_zoom
  n_summaries_kalenderns_detaljnivaer_och_lane_packning --> n_notes_kalenderns_zoom_styrs_av_vertikal_scrollning
  n_summaries_llm_wiki_karpathy --> n_notes_artikeln_om_agentic_software_factory_vad_som_ar_relevant_for_agentsystemet
  n_summaries_llm_wiki_karpathy --> n_notes_kunskaps_wiki
  n_summaries_llm_wiki_karpathy --> n_notes_kureringspasset_hittade_foraldrad_kunskap_i_wikin
  n_summaries_llm_wiki_karpathy --> n_notes_utvecklingsflode
  n_summaries_longhorn_disk_guide --> n_entities_talos_linux
  n_summaries_longhorn_disk_guide --> n_concepts_flux_operator
  n_summaries_longhorn_disk_guide --> n_concepts_longhorn
```

## Backlinks

- **concepts/authentik** <- [[concepts/oidc]], [[concepts/zitadel]]
- **concepts/cert-manager** <- [[concepts/envoy-gateway]], [[concepts/http-routing-flow]], [[entities/kubernetes]]
- **concepts/cloudflare-tunnel** <- [[concepts/http-routing-flow]]
- **concepts/cloudnative-pg** <- [[concepts/authentik]], [[concepts/zitadel]], [[entities/kubernetes]]
- **concepts/envoy-gateway** <- [[concepts/http-routing-flow]], [[concepts/kube-prometheus]], [[concepts/metallb]], [[concepts/oidc]], [[concepts/zitadel]], [[entities/kubernetes]]
- **concepts/flux-operator** <- [[concepts/gitops]], [[concepts/kyverno-policies]], [[concepts/sops]], [[entities/flux-gitops]], [[entities/kubernetes]], [[entities/talos-linux]], [[summaries/flux-components]], [[summaries/infrastructure-flux]], [[summaries/longhorn-disk-guide]]
- **concepts/gitops** <- [[concepts/cert-manager]], [[concepts/flux-operator]], [[entities/flux-gitops]], [[entities/kubernetes]], [[entities/talos-linux]], [[notes/infrastructure-agents-md-uppdaterad-tre-fallor-for-agenter]], [[summaries/infrastructure-flux]]
- **concepts/http-routing-flow** <- [[concepts/cloudflare-tunnel]], [[concepts/envoy-gateway]], [[notes/ipv6-ingress-till-envoy-designlage-2026-11-15]]
- **concepts/kube-prometheus** <- [[entities/kubernetes]]
- **concepts/kubelet-csr-approver** <- [[entities/kubernetes]]
- **concepts/kyverno-cel** <- [[concepts/kyverno-policies]]
- **concepts/kyverno-policies** <- [[concepts/kyverno-cel]], [[entities/kubernetes]]
- **concepts/longhorn** <- [[concepts/cloudnative-pg]], [[entities/kubernetes]], [[notes/usb-disken-acer-wd-320gb-status-och-longhorn-bedomning]], [[summaries/longhorn-disk-guide]]
- **concepts/metallb** <- [[concepts/envoy-gateway]], [[concepts/http-routing-flow]], [[entities/kubernetes]]
- **concepts/neon-cli-var-neon-och-neon-ts-maste-ligga** <- [[notes/pappa-kor-nu-pa-neon-neon-ts-env-local-och-poolad-direkt-wiring]]
- **concepts/oidc** <- [[concepts/authentik]], [[concepts/envoy-gateway]], [[concepts/zitadel]], [[entities/kubernetes]]
- **concepts/poolad-eller-direkt-anslutning-mot-neon** <- [[notes/pappa-kor-nu-pa-neon-neon-ts-env-local-och-poolad-direkt-wiring]]
- **concepts/sops** <- [[concepts/cloudflare-tunnel]], [[concepts/flux-operator]], [[concepts/gitops]], [[entities/flux-gitops]], [[entities/kubernetes]], [[entities/talos-linux]], [[notes/infrastructure-agents-md-uppdaterad-tre-fallor-for-agenter]]
- **concepts/zitadel** <- [[concepts/authentik]], [[concepts/cloudnative-pg]], [[concepts/envoy-gateway]], [[concepts/oidc]], [[entities/kubernetes]], [[notes/authflytten-i-plan-sessionen-ligger-nu-i-go-api-t-inte-i-nuxt]]
- **entities/flux-gitops** <- [[concepts/flux-operator]], [[concepts/gitops]], [[concepts/sops]], [[entities/kubernetes]], [[entities/talos-linux]], [[summaries/flux-components]]
- **entities/hdmi-kvm-switch-16-port-labbet** <- [[notes/kvm-styrning-fran-pxe-pi-n-16-port-hdmi-kvm]], [[notes/simon-kvm-cli-kommando-for-att-byta-hdmi-kvm-port-fran-pxe-vard]], [[summaries/16x1-hdmi-kvm-rs232-lan-protokoll]]
- **entities/kubernetes** <- [[concepts/authentik]], [[concepts/cert-manager]], [[concepts/cloudflare-tunnel]], [[concepts/cloudnative-pg]], [[concepts/envoy-gateway]], [[concepts/flux-operator]], [[concepts/gitops]], [[concepts/kube-prometheus]], [[concepts/kubelet-csr-approver]], [[concepts/kyverno-cel]], [[concepts/kyverno-policies]], [[concepts/longhorn]], [[concepts/metallb]], [[concepts/oidc]], [[concepts/sops]], [[concepts/zitadel]], [[entities/flux-gitops]], [[entities/pxe-server]], [[entities/talos-linux]], [[summaries/flux-components]]
- **entities/pxe-server** <- [[entities/kubernetes]], [[entities/talos-linux]], [[notes/bash-fallgropar-i-pxe-flash-tab-som-ifs-whitespace-och-read-med-for-fa-variabler]], [[notes/controlplane-recovery-via-pxe-ar-inte-konfigurerad]], [[notes/dnsmasq-dhcp-boot-tar-bara-en-tagg-anvand-tag-if-for-villkor]], [[notes/gotcha-pxe-forst-gor-noden-till-en-pxe-nod-maintenance-mode]], [[notes/korrigering-rock4c-plus-bin-ar-ett-microsd-emmc-image-inte-en-spi-image]], [[notes/mellanslag-i-radxa-rock-4c-saboterar-byggen-losning-bind-mount]], [[notes/pa-rock-plattorna-finns-ingen-bootloader-i-hardvaran]], [[notes/pxe-aterstallning-nodspecifika-matchbox-profiler-och-maintenance-mode-flode]], [[notes/pxe-boot-script-dhcp-kraver-mkimage-uimage-inte-ren-hush-text]], [[notes/pxe-forst-bootloader-for-rock-4c-byggd-som-patchad-u-boot-image]], [[notes/pxe-genererade-nodconfigs-var-stale-efter-talos-v1-14-uppgraderingen]], [[notes/pxe-kedjan-dhcp-boot-pekar-pa-textskript-matchbox-efi-ar-uefi-vagen]], [[notes/pxe-kedjan-for-worker-1-verifierad-serverside-pxe-metoden-failar-snabbt-utan-pxelinux-cfg]], [[notes/pxe-mainline-u-boots-bootefi-kraver-addr-size-for-image-och-initrd]], [[notes/pxe-status-fungerar-i-produktionsklustret-tftp-svarar-inte-fran-pod-natet]], [[notes/pxe-status-pa-dashboard-startsidan-tftp-gar-inte-att-kontrollera-fran-pod-natet]], [[notes/pxe-varden-korde-foraldrad-matchbox-nodconfiger-saknades-404]], [[notes/rk3399-s-bootrom-soker-spi-emmc-sd-usb-hardkodat]], [[notes/rock-4c-och-4se-har-ingen-spi-flash-i-fabrik-u-boot-ligger-pa-microsd-emmc]], [[notes/rock-pi-4-rk3399-defconfig-bygger-en-rock-pi-4a-for-4c-ska-rock-4c-plus-rk3399-defconfig-anvandas]], [[notes/simon-pxe-flash-clean-code-omgang-2-fokuserade-funktioner]], [[notes/simon-pxe-flash-flashar-pxe-forst-u-boot-till-sd-kort-i-pxe-servern]], [[notes/tre-satt-att-pxe-boota-rock-4c-och-varfor-dhcp-metoden-valdes]], [[notes/u-boot-2026-01-rock-4c-saknar-bootfile-boot-script-dhcp-natverksboot-dor-direkt]], [[notes/u-boot-for-rock-4c-built-from-mainline-with-pxe-first]], [[notes/u-boot-hamtar-alltid-boot-script-dhcp-per-nod-config-via-ipaddr]], [[notes/u-boot-kors-en-gang-boot-targets-styr-steg-2-inte-steg-1]], [[notes/worker-1-kortet-flashat-med-ny-mainline-build-tf-a-fran-repot]], [[notes/worker-1-och-worker-2-ar-doda-pa-link-niva-pxe-kedjan-ar-inte-blockerande]], [[notes/worker-1-rock-4c-nar-pxe-via-u-boot-boot-targets-men-pxe-ligger-sist]], [[notes/worker-1-rock-4c-tyst-mainline-u-boot-byggd-men-boarden-laser-inte-kortet]], [[notes/worker-1-startar-om-med-strom-men-skickar-noll-paket-fysisk-omstart-bevisar-samma-fel]], [[notes/worker-1-svarar-inte-ens-med-en-full-poe-cykel-noll-paket-i-bade-riktningar]], [[notes/worker-1-svarar-inte-i-natet-felsokningsstatus]], [[notes/worker-2-utan-strom-port-22-up-false-poe-power-0]], [[notes/workers-bootar-inte-via-pxe-trots-fungerande-pxe-server-unifi-poe-styrning-via-rest-device]]
- **entities/talos-linux** <- [[concepts/cert-manager]], [[concepts/kubelet-csr-approver]], [[concepts/longhorn]], [[concepts/sops]], [[entities/flux-gitops]], [[entities/kubernetes]], [[entities/pxe-server]], [[notes/pxe-genererade-nodconfigs-var-stale-efter-talos-v1-14-uppgraderingen]], [[notes/talos-v1-14-machineconfig-specen-ar-en-quoted-string-och-flera-dokument]], [[notes/usb-disken-acer-wd-320gb-status-och-longhorn-bedomning]], [[summaries/longhorn-disk-guide]]
- **notes/accuracy-fel-omstart** <- [[notes/session-livscykel]]
- **notes/admin-inloggning-via-losenord-i-stallet-for-token-i-ui-t** <- [[notes/pappa-admin-token-ett-stalle-for-token-inte-handpallagning-per-miljo]]
- **notes/agent-api-crashade-i-tilt-saknad-database-url-mot-cnpg-databasen-i-agents-dev** <- [[notes/agent-granssnittet-nuxt-saknades-helt-i-tilt-routes-forsvann-i-refaktoreringen-677b876]]
- **notes/agent-granssnittet-nuxt-saknades-helt-i-tilt-routes-forsvann-i-refaktoreringen-677b876** <- [[notes/agent-api-crashade-i-tilt-saknad-database-url-mot-cnpg-databasen-i-agents-dev]]
- **notes/arbetsstationen-ar-omarchy-arch-apt-get-finns-inte-anvand-pacman-yay** <- [[notes/utvecklingsflode]]
- **notes/arkitektur** <- [[notes/bigram-algoritmen]], [[notes/konventioner]], [[notes/session-livscykel]], [[notes/utvecklingsflode]]
- **notes/artikeln-om-agentic-software-factory-vad-som-ar-relevant-for-agentsystemet** <- [[notes/verifierade-policy-hal-i-tool-lagret-kommandon-ar-blocklist-inte-allowlist]], [[summaries/llm-wiki-karpathy]]
- **notes/aterkommande-serier-doljs-inte-langre-beroende-pa-zoomniva** <- [[notes/kalendern-begransar-aterkommande-forekomster-i-langtidvyer]], [[notes/uttunnade-serier-vags-sa-att-antalen-blir-exakta]]
- **notes/authflytten-i-plan-sessionen-ligger-nu-i-go-api-t-inte-i-nuxt** <- [[notes/plan-repots-ci-och-prod-pipeline-var-trasig-fixade-sokvagar-och-bygg]]
- **notes/automatisk-malikon-behovde-sokord-och-svenska-alias** <- [[notes/nya-mal-far-ikon-fran-titeln]]
- **notes/bash-fallgropar-i-pxe-flash-tab-som-ifs-whitespace-och-read-med-for-fa-variabler** <- [[entities/pxe-server]]
- **notes/beslut-pi-durable-satts-pa-hyllan-for-tillfallet** <- [[notes/pi-durable-1-4-0-api-avvikelser-fran-blogginlagget-och-tva-gotchas]], [[notes/pi-durable-demo-paketet-earendil-works-pi-durable-fungerar]], [[notes/pi-durable-fungerar-med-minimax-coding-plan-anthropic-messages-api]]
- **notes/bigram-algoritmen** <- [[notes/arkitektur]], [[notes/wordlist-kallor]]
- **notes/boka-enskild-kommunbokning-i-matchi-fran-bookings** <- [[notes/matchi-bokning-slot-id-andras-nar-bokningen-skapas]]
- **notes/cachade-imap-sessioner-kraver-hantering-av-doda-anslutningar** <- [[notes/mailmaster-kontodropdown-outlook-inloggning-via-device-code]]
- **notes/ci-saknade-inloggning-mot-ghcr-alla-image-pushar-misslyckades** <- [[notes/pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres]]
- **notes/cluster1-stale-kubeconfig-ca-och-noll-workers-diagnos-2026-11-11** <- [[notes/worker-7-s-kubeletcertifikat-signerades-av-en-gammal-kubernetes-ca-som-finns-kvar-i-git-historiken]], [[notes/worker-8-korrupt-xfs-pa-ephemeral-partitionen-gjorde-att-kubelet-aldrig-startade-fix-med-mkfs-xfs]]
- **notes/controlplane-recovery-via-pxe-ar-inte-konfigurerad** <- [[entities/pxe-server]]
- **notes/cookies-json-var-versionsstyrd-i-offentligt-neuro-repo** <- [[notes/neuroquant-vikterna-raknas-om-dagligen-mejl-triggas-pa-publicering]]
- **notes/detaljvy-for-kandidat-i-pappa-gemensamma-vagar-mellan-kandidater** <- [[notes/redigerare-och-agarskap-for-relationer-i-pappa]], [[notes/relationer-mellan-kandidater-i-pappa-symmetriska-par-med-styrka-1-10]]
- **notes/dev-inloggning-i-plan-session-token-db-rattigheter-och-oauth-url-er** <- [[notes/authflytten-i-plan-sessionen-ligger-nu-i-go-api-t-inte-i-nuxt]]
- **notes/dnsmasq-dhcp-boot-tar-bara-en-tagg-anvand-tag-if-for-villkor** <- [[entities/pxe-server]]
- **notes/en-talos-efi-partition-pa-kortet-blockerar-pxe-booten-bootflow-loaded-fore-natverket** <- [[notes/radxa-bradena-behover-bootmedium-men-talos-och-bootloader-maste-ligga-pa-skilda-medium]]
- **notes/en-uttunnad-forekomst-maste-vara-synlig-aven-om-den-ligger-utanfor-fonstret** <- [[notes/aterkommande-serier-doljs-inte-langre-beroende-pa-zoomniva]]
- **notes/enter-pa-goals-skapar-nytt-grundmal-istallet-for-att-oppna-malet** <- [[notes/mal-kan-ha-en-lang-beskrivning-redigerad-med-tangenten-t]]
- **notes/erfarenhetsfilter-i-pappa-it-experience-years-med-sparbart-grund** <- [[notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen]], [[notes/pappa-anstallningshistorik-i-cv-format-fix-av-kompetensordboken]]
- **notes/eviga-mal-i-plan-kan-inte-slutforas** <- [[notes/enter-pa-goals-skapar-nytt-grundmal-istallet-for-att-oppna-malet]], [[notes/mal-kan-ha-en-lang-beskrivning-redigerad-med-tangenten-t]]
- **notes/expanderade-eml-laggs-in-som-riktiga-mejl-i-mappen-append** <- [[notes/mailagenten-expanderar-bifogade-eml-till-egna-mejl]], [[notes/moderposten-arkiveras-nar-eml-bilagorna-har-lagts-in]]
- **notes/flera-kalendrar-med-egna-farger-i-plan** <- [[notes/handelsen-har-ingen-egen-farg]], [[notes/standardkalender-i-plan]], [[notes/valfri-fargkod-pa-kalendrar]], [[notes/vertikal-kalender-med-ikonstod-i-plan]]
- **notes/globala-agentinstruktioner-deployas-via-dotfiles** <- [[notes/kunskaps-wiki]]
- **notes/go-integrationstest-tomde-kalenderhandelserna-i-neon-igen** <- [[notes/kalenderhandelser-aterstallda-fran-neon-wal-historik]]
- **notes/goals-listar-bara-grundmal-scroll-vid-j-k** <- [[notes/enter-pa-goals-skapar-nytt-grundmal-istallet-for-att-oppna-malet]]
- **notes/gotcha-pxe-forst-gor-noden-till-en-pxe-nod-maintenance-mode** <- [[entities/pxe-server]], [[notes/radxa-bradena-behover-bootmedium-men-talos-och-bootloader-maste-ligga-pa-skilda-medium]]
- **notes/gron-markering-av-tider-som-ar-bokade-i-matchi-skrivlasande-jamforelse** <- [[notes/synk-hallbokningar-kommunens-actor-smartbook-matchi]]
- **notes/ha-automationer-behover-en-startv-g-for-att-aterstalla-laget-och-tre-syntaxfallor** <- [[notes/skogsduvan-temperaturkaskad-i-tre-steg-styrd-av-minsta-temperaturen]]
- **notes/handelsen-har-ingen-egen-farg** <- [[notes/flera-kalendrar-med-egna-farger-i-plan]]
- **notes/ikonvaljaren-i-plan-api-icons-routen-var-borttagen-och-ikonprovider-none-gjorde-att-inga-ikoner-renderades** <- [[notes/vertikal-kalender-med-ikonstod-i-plan]]
- **notes/imap-inloggning-for-michael-brundin-soulmate-se-kraver-admin-consent** <- [[notes/mailmaster-kontodropdown-outlook-inloggning-via-device-code]]
- **notes/infrastructure-agents-md-uppdaterad-tre-fallor-for-agenter** <- [[notes/kunskaps-wiki]]
- **notes/ipv6-ingress-till-envoy-designlage-2026-11-15** <- [[notes/unifi-gatewayen-har-ingen-publik-ipv6-cgnat-via-starlink]]
- **notes/it-arbetsplatser-i-sundsvall-3-469-enligt-scb-yrkesregistret** <- [[notes/konsulter-i-sundsvall-antalet-gar-inte-att-mata-det-som-gar-att-saga]], [[notes/konsulter-vs-anstallda-i-sundsvall-halva-it-styrkan-sitter-i-konsultnara-bolag]], [[notes/marknadslage-panel-hogst-upp-pa-personer-i-pappa]], [[notes/rekrytering-av-it-personal-i-sundsvall-marknadsbild-och-frolista]]
- **notes/kalenderhandelser-aterstallda-fran-neon-wal-historik** <- [[notes/go-integrationstest-tomde-kalenderhandelserna-i-neon-igen]]
- **notes/kalendern-hamtar-alla-handelser-en-gang-och-expanderar-serier-lokalt** <- [[notes/kalendern-begransar-aterkommande-forekomster-i-langtidvyer]], [[notes/lokal-expansion-kan-verifieras-mot-serverns-expansion]], [[notes/vertikal-kalender-med-kontinuerlig-zoom]]
- **notes/kalenderns-detaljnivaer-fem-nivaer-och-lane-packning** <- [[notes/kalenderns-nivavaxling-animeras-crossfade-pop-in-och-hysteresis]], [[notes/vertikal-kalender-med-ikonstod-i-plan]], [[summaries/kalenderns-detaljnivaer-och-lane-packning]]
- **notes/kalenderns-nivavaxling-animeras-crossfade-pop-in-och-hysteresis** <- [[notes/kalenderns-detaljnivaer-fem-nivaer-och-lane-packning]], [[notes/kalenderns-tidsaxelrader-animeras-vid-zoom]], [[notes/parallella-agent-sessioner-i-samma-repo-och-databas]], [[notes/vertikal-kalender-med-kontinuerlig-zoom]]
- **notes/kalenderns-tidsaxelrader-animeras-vid-zoom** <- [[notes/kalenderns-nivavaxling-animeras-crossfade-pop-in-och-hysteresis]]
- **notes/kalenderns-zoom-styrs-av-vertikal-scrollning** <- [[notes/vertikal-kalender-med-ikonstod-i-plan]], [[summaries/kalenderns-detaljnivaer-och-lane-packning]]
- **notes/kameo-beloppet-kommer-fran-transaktionerna-inte-fran-mejlet** <- [[notes/kameo-utbetalningar-nuxt-sida-som-laser-lan-fran-kameo]]
- **notes/kameo-ranteutbetalning-redan-utbetald-samlingsuttag-pa-47-29-kr** <- [[notes/kameo-utbetalningar-nuxt-sida-som-laser-lan-fran-kameo]], [[notes/rantebetalningsmejl-hamnar-i-inbox-deleted-messages-och-arkivet-matchades-aldrig]]
- **notes/kameo-sidan-flyttad-till-toppniva-och-delad-i-flikar** <- [[notes/kameo-utbetalningar-nuxt-sida-som-laser-lan-fran-kameo]]
- **notes/kameo-uppgifter-synkas-till-fil-sa-docker-slipper-op-cli** <- [[notes/kameo-utbetalningar-nuxt-sida-som-laser-lan-fran-kameo]]
- **notes/kameo-utbetalningar-nuxt-sida-som-laser-lan-fran-kameo** <- [[notes/kameo-beloppet-kommer-fran-transaktionerna-inte-fran-mejlet]]
- **notes/kanata-switch-session-tmuxinator-fonster-oppnades-inte** <- [[notes/arbetsstationen-ar-omarchy-arch-apt-get-finns-inte-anvand-pacman-yay]]
- **notes/kandidatlistan-heter-nu-personer** <- [[notes/nuxt-sidomeny-kandidater-som-startsida-i-pappa]]
- **notes/klustret-anvander-sops-via-flux-inte-vault-eso** <- [[concepts/sops]], [[entities/kubernetes]]
- **notes/klustret-har-byggts-om-bara-kube-system-flux-och-appar-saknas-2026-11-23** <- [[entities/kubernetes]]
- **notes/konsulter-i-sundsvall-antalet-gar-inte-att-mata-det-som-gar-att-saga** <- [[notes/konsulter-vs-anstallda-i-sundsvall-halva-it-styrkan-sitter-i-konsultnara-bolag]]
- **notes/konventioner** <- [[notes/kunskaps-wiki]], [[notes/utvecklingsflode]]
- **notes/korrigering-rock4c-plus-bin-ar-ett-microsd-emmc-image-inte-en-spi-image** <- [[entities/pxe-server]]
- **notes/krypterade-secrets-i-klustret-ogiltiga-inloggningar-och-keycloak-brute-force** <- [[notes/synk-hallbokningar-kommunens-actor-smartbook-matchi]]
- **notes/kunskaps-wiki** <- [[notes/arkitektur]], [[notes/stray-wiki-i-typetrainer-repot-ar-inte-kallan-for-wikin]], [[notes/utvecklingsflode]], [[summaries/llm-wiki-karpathy]]
- **notes/kureringspasset-hittade-foraldrad-kunskap-i-wikin** <- [[notes/kunskaps-wiki]], [[summaries/llm-wiki-karpathy]]
- **notes/kvm-portmappning-worker-n-port-n-den-deployade-configen-pa-pi-n-var-fel** <- [[notes/simon-kvm-cli-kommando-for-att-byta-hdmi-kvm-port-fran-pxe-vard]]
- **notes/kvm-styrning-fran-pxe-pi-n-16-port-hdmi-kvm** <- [[entities/hdmi-kvm-switch-16-port-labbet]], [[notes/simon-kvm-cli-kommando-for-att-byta-hdmi-kvm-port-fran-pxe-vard]], [[summaries/16x1-hdmi-kvm-rs232-lan-protokoll]]
- **notes/leader-key-gjordes-global-mellanslag-sekvenser-pa-alla-sidor** <- [[notes/mellanslag-blockerades-i-malbeskrivningen]]
- **notes/learn-calibre-synk-och-automatisk-import-via-systemd** <- [[notes/learn-kunskapssystem-implementerat-i-src-python]]
- **notes/learn-designsystem-a3-folio-i-nuxt** <- [[notes/learn-rest-api-och-nuxt-spa-mot-samma-kunskapsbas]]
- **notes/learn-korsanalys-av-varje-bok-mot-den-samlade-kunskapsbasen** <- [[notes/learn-kunskapsbasen-kan-lagras-i-postgresql]]
- **notes/learn-kunskapsbasen-kan-lagras-i-postgresql** <- [[notes/learn-calibre-synk-och-automatisk-import-via-systemd]], [[notes/learn-korsanalys-av-varje-bok-mot-den-samlade-kunskapsbasen]]
- **notes/learn-kunskapssystem-implementerat-i-src-python** <- [[notes/learn-kunskapssystemet-portat-fran-python-till-go]]
- **notes/learn-kunskapssystemet-portat-fran-python-till-go** <- [[notes/learn-kunskapssystem-implementerat-i-src-python]], [[notes/learn-rest-api-och-nuxt-spa-mot-samma-kunskapsbas]]
- **notes/learn-rest-api-och-nuxt-spa-mot-samma-kunskapsbas** <- [[notes/learn-kunskapssystemet-portat-fran-python-till-go]]
- **notes/lightcraft-photocraft-installerade-pa-omarchy-workstation-utan-sudo** <- [[notes/photocraft-kamerans-look-varning-vid-raw-oppning-och-hur-den-hanteras]]
- **notes/losning-1password-backed-env-variabel-for-mcp-tokens** <- [[notes/1password-mcp-environments-pa-arch-omarchy-kraver-setgid-pa-1password-mcp]], [[notes/neon-api-token-hamnade-i-dotfiles-git-historiken]]
- **notes/lunchmenyn-delar-en-modell-mellan-byggskript-och-webblasare** <- [[notes/nuxt-4-webbapp-for-pappa-laser-personerna-ur-neon]]
- **notes/mailagent-bygget-brot-pa-ofardig-kameo-integration** <- [[notes/mailagenten-flyttad-fran-mailmaster-repot-in-i-pappa]]
- **notes/mailagenten-expanderar-bifogade-eml-till-egna-mejl** <- [[notes/expanderade-eml-laggs-in-som-riktiga-mejl-i-mappen-append]]
- **notes/mailagenten-flyttad-fran-mailmaster-repot-in-i-pappa** <- [[notes/mailagentens-mcp-verktygen-saknade-konto-och-hade-fel-schema]]
- **notes/mailagentens-kontodropdown-renderades-aldrig-nuxt-prefixade-komponenter** <- [[notes/mailagentens-kontofel-gav-alltid-http-500-och-listan-foljde-inte-kontot]], [[notes/mailagentens-mcp-verktygen-saknade-konto-och-hade-fel-schema]]
- **notes/mailagentens-kontofel-gav-alltid-http-500-och-listan-foljde-inte-kontot** <- [[notes/mailagentens-kontodropdown-renderades-aldrig-nuxt-prefixade-komponenter]]
- **notes/mailagentens-mcp-verktygen-saknade-konto-och-hade-fel-schema** <- [[notes/mailagentens-kontodropdown-renderades-aldrig-nuxt-prefixade-komponenter]], [[notes/mailagentens-kontofel-gav-alltid-http-500-och-listan-foljde-inte-kontot]]
- **notes/mailmaster-api-v1-gav-failed-to-fetch-go-servern-panickade-och-saknade-cors** <- [[notes/mailagentens-kontodropdown-renderades-aldrig-nuxt-prefixade-komponenter]]
- **notes/mailmaster-gammal-container-build-korde-process-och-arkiverade-mejl** <- [[notes/mailagenten-flyttad-fran-mailmaster-repot-in-i-pappa]]
- **notes/mailmaster-granskning-av-ai-atgarder-fore-utforande** <- [[notes/mailmaster-svarmejl-fick-fel-brodtext-tre-buggar-i-svarttextskedjan]]
- **notes/mailmaster-kontodropdown-outlook-inloggning-via-device-code** <- [[notes/cachade-imap-sessioner-kraver-hantering-av-doda-anslutningar]], [[notes/imap-inloggning-for-michael-brundin-soulmate-se-kraver-admin-consent]]
- **notes/mailmaster-svarmejl-fick-fel-brodtext-tre-buggar-i-svarttextskedjan** <- [[notes/mailagentens-mcp-verktygen-saknade-konto-och-hade-fel-schema]]
- **notes/mailvyn-mail-omstegad-till-tailwind-tema** <- [[notes/nuxt-4-webbapp-for-pappa-laser-personerna-ur-neon]], [[notes/plan-appen-morkt-designsystem-med-semantiska-farg-token]]
- **notes/mal-kan-ha-due-datum-deadline** <- [[notes/eviga-mal-i-plan-kan-inte-slutforas]], [[notes/mal-kan-ha-startdatum-syns-i-priority-forst-nar-datumet-passerat]], [[notes/manuell-sortering-av-mal-pa-priority]], [[notes/priorityfilter-ska-behandla-saknat-started-som-ej-paborjat]]
- **notes/mal-kan-ha-en-lang-beskrivning-redigerad-med-tangenten-t** <- [[notes/eviga-mal-i-plan-kan-inte-slutforas]]
- **notes/mal-kan-ha-startdatum-syns-i-priority-forst-nar-datumet-passerat** <- [[notes/mal-kan-ha-due-datum-deadline]]
- **notes/malgraph-navigering-mellan-malsidor-gar-fran-sekunder-till-millisekunder** <- [[notes/platser-pa-mal-och-kalenderhandelser-openspec-change-add-places]]
- **notes/malikoner-anvander-enkel-oversattning-fore-iconify** <- [[notes/nya-mal-far-ikon-fran-titeln]]
- **notes/manuell-sortering-av-mal-pa-priority** <- [[notes/eviga-mal-i-plan-kan-inte-slutforas]]
- **notes/marknadslage-panel-hogst-upp-pa-personer-i-pappa** <- [[notes/konsulter-i-sundsvall-antalet-gar-inte-att-mata-det-som-gar-att-saga]], [[notes/konsulter-vs-anstallda-i-sundsvall-halva-it-styrkan-sitter-i-konsultnara-bolag]]
- **notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen** <- [[concepts/poolad-eller-direkt-anslutning-mot-neon]], [[notes/erfarenhetsfilter-i-pappa-it-experience-years-med-sparbart-grund]], [[notes/nuxt-4-webbapp-for-pappa-laser-personerna-ur-neon]], [[notes/pappa-anstallningshistorik-i-cv-format-fix-av-kompetensordboken]], [[notes/pappa-dubblettknapp-mansklig-granskning-och-bestandig-sammanslagning]], [[notes/pappa-uppdragsmatchning-delpoang-forklarande-ranking]], [[notes/pappa-yaml-frolistan-kan-aterskapas-ur-masterdatabasen]], [[notes/rekrytering-av-it-personal-i-sundsvall-marknadsbild-och-frolista]]
- **notes/matchi-bokning-slot-id-andras-nar-bokningen-skapas** <- [[notes/synk-hallbokningar-kommunens-actor-smartbook-matchi]]
- **notes/matchi-inloggningen-ar-flaky-klienten-forsoker-om-och-dev-klustret-behover-ombyggnad** <- [[notes/matchi-keycloak-session-code-maste-fangas-med-generell-teckenklass]]
- **notes/matchi-keycloak-session-code-maste-fangas-med-generell-teckenklass** <- [[notes/synk-hallbokningar-kommunens-actor-smartbook-matchi]]
- **notes/md-viewer-tauri-installerades-men-avinstallerades-igen** <- [[notes/beslut-pi-durable-satts-pa-hyllan-for-tillfallet]]
- **notes/median-i-stallet-for-snitt-for-wpm** <- [[notes/accuracy-fel-omstart]]
- **notes/mellanslag-blockerades-i-malbeskrivningen** <- [[notes/leader-key-gjordes-global-mellanslag-sekvenser-pa-alla-sidor]], [[notes/mal-kan-ha-en-lang-beskrivning-redigerad-med-tangenten-t]]
- **notes/mellanslag-i-radxa-rock-4c-saboterar-byggen-losning-bind-mount** <- [[entities/pxe-server]]
- **notes/michael-brundin-blir-en-personnod-i-relationsgrafen** <- [[notes/relationsgrafen-anvander-fast-michael-id-utan-person-kind]]
- **notes/min-kontaktlista-jag-som-nod-i-grafen-pappa** <- [[notes/redigerare-och-agarskap-for-relationer-i-pappa]]
- **notes/moderposten-arkiveras-nar-eml-bilagorna-har-lagts-in** <- [[notes/expanderade-eml-laggs-in-som-riktiga-mejl-i-mappen-append]]
- **notes/neon-api-token-hamnade-i-dotfiles-git-historiken** <- [[notes/cookies-json-var-versionsstyrd-i-offentligt-neuro-repo]], [[notes/losning-1password-backed-env-variabel-for-mcp-tokens]]
- **notes/neuroquant-ar-ombyggt-trpc-api-och-e-post-verifiering** <- [[notes/neuroquant-kontot-last-av-for-manga-inloggningsforsok]], [[notes/neuroquant-vikterna-raknas-om-dagligen-mejl-triggas-pa-publicering]]
- **notes/neuroquant-kontot-last-av-for-manga-inloggningsforsok** <- [[notes/neuroquant-ar-ombyggt-trpc-api-och-e-post-verifiering]], [[notes/neuroquant-vikterna-raknas-om-dagligen-mejl-triggas-pa-publicering]]
- **notes/neuroquant-vikterna-raknas-om-dagligen-mejl-triggas-pa-publicering** <- [[notes/cookies-json-var-versionsstyrd-i-offentligt-neuro-repo]], [[notes/neuroquant-ar-ombyggt-trpc-api-och-e-post-verifiering]], [[notes/neuroquant-kontot-last-av-for-manga-inloggningsforsok]]
- **notes/nuxt-4-webbapp-for-pappa-laser-personerna-ur-neon** <- [[notes/erfarenhetsfilter-i-pappa-it-experience-years-med-sparbart-grund]], [[notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen]], [[notes/pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres]]
- **notes/nuxt-cli-via-bun-run-dev-gar-sonder-i-stk** <- [[notes/nuxt-4-webbapp-for-pappa-laser-personerna-ur-neon]]
- **notes/nuxt-sidomeny-kandidater-som-startsida-i-pappa** <- [[notes/kandidatlistan-heter-nu-personer]], [[notes/nuxt-4-webbapp-for-pappa-laser-personerna-ur-neon]]
- **notes/nya-mal-far-ikon-fran-titeln** <- [[notes/vertikal-kalender-med-ikonstod-i-plan]]
- **notes/one-com-lagrar-mappar-i-nfd-normalisera-i-getfoldername** <- [[notes/mailmaster-kontodropdown-outlook-inloggning-via-device-code]]
- **notes/pa-rock-plattorna-finns-ingen-bootloader-i-hardvaran** <- [[entities/pxe-server]], [[notes/radxa-bradena-behover-bootmedium-men-talos-och-bootloader-maste-ligga-pa-skilda-medium]], [[notes/rk3399-s-bootrom-soker-spi-emmc-sd-usb-hardkodat]], [[notes/u-boot-2026-01-rock-4c-saknar-bootfile-boot-script-dhcp-natverksboot-dor-direkt]]
- **notes/pappa-admin-token-ett-stalle-for-token-inte-handpallagning-per-miljo** <- [[notes/pappa-hemlighet-i-build-context-env-i-docker-imagen]]
- **notes/pappa-anstallningshistorik-i-cv-format-fix-av-kompetensordboken** <- [[notes/erfarenhetsfilter-i-pappa-it-experience-years-med-sparbart-grund]], [[notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen]], [[notes/pappa-yaml-frolistan-kan-aterskapas-ur-masterdatabasen]], [[notes/rekrytering-av-it-personal-i-sundsvall-marknadsbild-och-frolista]], [[notes/utbyggnad-av-pappas-personlista-bolagsregister-konferensbios-gav-140-namngivna-personer]]
- **notes/pappa-dubblettknapp-mansklig-granskning-och-bestandig-sammanslagning** <- [[notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen]], [[notes/pappa-admin-token-ett-stalle-for-token-inte-handpallagning-per-miljo]], [[notes/pappa-uppdragsmatchning-delpoang-forklarande-ranking]], [[notes/pappa-vyn-uppdrag-kravformular-och-rankad-kandidatlista]], [[notes/pappa-yaml-frolistan-kan-aterskapas-ur-masterdatabasen]], [[notes/personer-kan-laggas-in-manuellt-i-pappa]], [[notes/relationer-mellan-kandidater-i-pappa-symmetriska-par-med-styrka-1-10]]
- **notes/pappa-hemlighet-i-build-context-env-i-docker-imagen** <- [[notes/pappa-admin-token-ett-stalle-for-token-inte-handpallagning-per-miljo]]
- **notes/pappa-kor-nu-pa-neon-neon-ts-env-local-och-poolad-direkt-wiring** <- [[notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen]], [[notes/pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres]]
- **notes/pappa-regelendpointet-laser-och-sparar-riktig-rules-md-var-stub** <- [[notes/mailagenten-flyttad-fran-mailmaster-repot-in-i-pappa]]
- **notes/pappa-uppdragsmatchning-delpoang-forklarande-ranking** <- [[notes/pappa-dubblettknapp-mansklig-granskning-och-bestandig-sammanslagning]], [[notes/pappa-vyn-uppdrag-kravformular-och-rankad-kandidatlista]]
- **notes/pappa-vyn-uppdrag-kravformular-och-rankad-kandidatlista** <- [[notes/pappa-admin-token-ett-stalle-for-token-inte-handpallagning-per-miljo]]
- **notes/pappa-yaml-frolistan-kan-aterskapas-ur-masterdatabasen** <- [[notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen]]
- **notes/pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres** <- [[notes/nuxt-4-webbapp-for-pappa-laser-personerna-ur-neon]], [[notes/pappa-kor-nu-pa-neon-neon-ts-env-local-och-poolad-direkt-wiring]]
- **notes/pappas-grundare-ar-michael-brundin-inte-git-anvandaren-simon-brundin** <- [[notes/redigerare-och-agarskap-for-relationer-i-pappa]]
- **notes/parallella-agent-sessioner-i-samma-repo-och-databas** <- [[notes/go-integrationstest-tomde-kalenderhandelserna-i-neon-igen]], [[notes/vertikal-kalender-med-kontinuerlig-zoom]]
- **notes/personer-kan-laggas-in-manuellt-i-pappa** <- [[notes/pappa-dubblettknapp-mansklig-granskning-och-bestandig-sammanslagning]]
- **notes/photocraft-agentstyrning-via-control-protokoll-control-7878** <- [[notes/lightcraft-photocraft-installerade-pa-omarchy-workstation-utan-sudo]]
- **notes/photocraft-kamerans-look-varning-vid-raw-oppning-och-hur-den-hanteras** <- [[notes/photocraft-agentstyrning-via-control-protokoll-control-7878]]
- **notes/pi-desktop-notis-nar-agent-ar-klar** <- [[notes/beslut-pi-durable-satts-pa-hyllan-for-tillfallet]]
- **notes/pi-durable-1-4-0-api-avvikelser-fran-blogginlagget-och-tva-gotchas** <- [[notes/beslut-pi-durable-satts-pa-hyllan-for-tillfallet]], [[notes/pi-durable-demo-paketet-earendil-works-pi-durable-fungerar]], [[notes/pi-durable-fungerar-med-minimax-coding-plan-anthropic-messages-api]]
- **notes/pi-durable-demo-paketet-earendil-works-pi-durable-fungerar** <- [[notes/beslut-pi-durable-satts-pa-hyllan-for-tillfallet]], [[notes/pi-durable-1-4-0-api-avvikelser-fran-blogginlagget-och-tva-gotchas]], [[notes/pi-durable-fungerar-med-minimax-coding-plan-anthropic-messages-api]]
- **notes/pi-durable-fungerar-med-minimax-coding-plan-anthropic-messages-api** <- [[notes/beslut-pi-durable-satts-pa-hyllan-for-tillfallet]], [[notes/pi-durable-1-4-0-api-avvikelser-fran-blogginlagget-och-tva-gotchas]], [[notes/pi-durable-demo-paketet-earendil-works-pi-durable-fungerar]]
- **notes/pi-tog-bort-pi-rate-limit-och-pi-pattern-retry-fixade-omarchy-temasy** <- [[notes/arbetsstationen-ar-omarchy-arch-apt-get-finns-inte-anvand-pacman-yay]]
- **notes/picokvm-luckfox-startar-inte-i-labbet-lite-saknar-emmc-ingen-poe** <- [[notes/kvm-styrning-fran-pxe-pi-n-16-port-hdmi-kvm]]
- **notes/plan-api-anvander-scalar-for-openapi** <- [[notes/vertikal-kalender-med-ikonstod-i-plan]]
- **notes/plan-appen-morkt-designsystem-med-semantiska-farg-token** <- [[notes/mailvyn-mail-omstegad-till-tailwind-tema]]
- **notes/plan-cachad-spa-skal-gjorde-att-webblasaren-korde-gammalt-js-bundle-med-ikon-provider-none** <- [[notes/ikonvaljaren-i-plan-api-icons-routen-var-borttagen-och-ikonprovider-none-gjorde-att-inga-ikoner-renderades]]
- **notes/plan-dev-kor-mot-neon-istallet-for-lokal-postgres** <- [[notes/pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres]], [[notes/tidsatgang-pa-mal-i-plan-estimated-minutes]]
- **notes/plan-repots-ci-och-prod-pipeline-var-trasig-fixade-sokvagar-och-bygg** <- [[notes/pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres]]
- **notes/platser-pa-mal-och-kalenderhandelser-openspec-change-add-places** <- [[notes/malgraph-navigering-mellan-malsidor-gar-fran-sekunder-till-millisekunder]]
- **notes/popcorn-time-installerad-som-appimage-i-local** <- [[notes/qbittorrent-installerad-pa-workstation-omarchy-arch]]
- **notes/popcorn-time-remote-movies-api-failed-to-respond-orsak-och-fix** <- [[notes/popcorn-time-installerad-som-appimage-i-local]]
- **notes/popcorn-time-svenskt-ljud-finns-inte-i-katalogen-anvand-torrentsamling-torrent-collection** <- [[notes/popcorn-time-remote-movies-api-failed-to-respond-orsak-och-fix]]
- **notes/prioriteringsordningen-pa-now-sparades-inte-bulk-update-och-spar-ko** <- [[notes/sortering-av-grundmal-pa-goals-med-piltangenterna]]
- **notes/priorityfilter-ska-behandla-saknat-started-som-ej-paborjat** <- [[notes/eviga-mal-i-plan-kan-inte-slutforas]]
- **notes/pxe-aterstallning-nodspecifika-matchbox-profiler-och-maintenance-mode-flode** <- [[entities/pxe-server]]
- **notes/pxe-boot-script-dhcp-kraver-mkimage-uimage-inte-ren-hush-text** <- [[entities/pxe-server]]
- **notes/pxe-bootloop-initramfs-startar-om-for-att-boota-installerad-talos-pa-emmc** <- [[entities/pxe-server]]
- **notes/pxe-forst-bootloader-for-rock-4c-byggd-som-patchad-u-boot-image** <- [[entities/pxe-server]], [[notes/gotcha-pxe-forst-gor-noden-till-en-pxe-nod-maintenance-mode]], [[notes/pxe-kedjan-dhcp-boot-pekar-pa-textskript-matchbox-efi-ar-uefi-vagen]], [[notes/u-boot-kors-en-gang-boot-targets-styr-steg-2-inte-steg-1]], [[notes/worker-1-rock-4c-nar-pxe-via-u-boot-boot-targets-men-pxe-ligger-sist]]
- **notes/pxe-genererade-nodconfigs-var-stale-efter-talos-v1-14-uppgraderingen** <- [[entities/pxe-server]], [[notes/talos-v1-14-machineconfig-specen-ar-en-quoted-string-och-flera-dokument]]
- **notes/pxe-kedjan-dhcp-boot-pekar-pa-textskript-matchbox-efi-ar-uefi-vagen** <- [[entities/pxe-server]], [[notes/pxe-varden-korde-foraldrad-matchbox-nodconfiger-saknades-404]]
- **notes/pxe-kedjan-for-worker-1-verifierad-serverside-pxe-metoden-failar-snabbt-utan-pxelinux-cfg** <- [[entities/pxe-server]], [[notes/rock-pi-4-rk3399-defconfig-bygger-en-rock-pi-4a-for-4c-ska-rock-4c-plus-rk3399-defconfig-anvandas]], [[notes/u-boot-for-rock-4c-built-from-mainline-with-pxe-first]], [[notes/worker-1-kortet-flashat-med-ny-mainline-build-tf-a-fran-repot]]
- **notes/pxe-mainline-u-boots-bootefi-kraver-addr-size-for-image-och-initrd** <- [[entities/pxe-server]]
- **notes/pxe-status-fungerar-i-produktionsklustret-tftp-svarar-inte-fran-pod-natet** <- [[notes/pxe-status-pa-dashboard-startsidan-tftp-gar-inte-att-kontrollera-fran-pod-natet]]
- **notes/pxe-status-pa-dashboard-startsidan-tftp-gar-inte-att-kontrollera-fran-pod-natet** <- [[notes/pxe-status-fungerar-i-produktionsklustret-tftp-svarar-inte-fran-pod-natet]]
- **notes/pxe-varden-korde-foraldrad-matchbox-nodconfiger-saknades-404** <- [[entities/pxe-server]], [[notes/pxe-kedjan-dhcp-boot-pekar-pa-textskript-matchbox-efi-ar-uefi-vagen]]
- **notes/qbittorrent-ar-standardhanterare-for-torrent-och-magnet-lankar** <- [[notes/popcorn-time-svenskt-ljud-finns-inte-i-katalogen-anvand-torrentsamling-torrent-collection]]
- **notes/qbittorrent-installerad-pa-workstation-omarchy-arch** <- [[notes/qbittorrent-ar-standardhanterare-for-torrent-och-magnet-lankar]]
- **notes/query-parametrar-i-actor-smartbook-endpointen-url-avkodas-till-mellanslag** <- [[notes/synk-hallbokningar-kommunens-actor-smartbook-matchi]]
- **notes/rantebetalningsmejl-hamnar-i-inbox-deleted-messages-och-arkivet-matchades-aldrig** <- [[notes/kameo-beloppet-kommer-fran-transaktionerna-inte-fran-mejlet]]
- **notes/recovery-av-en-rock-worker-kraver-fysisk-rattning-esp-filen-pa-emmc-fangar-booten** <- [[notes/pxe-aterstallning-nodspecifika-matchbox-profiler-och-maintenance-mode-flode]]
- **notes/redigerare-och-agarskap-for-relationer-i-pappa** <- [[notes/detaljvy-for-kandidat-i-pappa-gemensamma-vagar-mellan-kandidater]], [[notes/min-kontaktlista-jag-som-nod-i-grafen-pappa]], [[notes/pappas-grundare-ar-michael-brundin-inte-git-anvandaren-simon-brundin]], [[notes/relationsoversikt-i-pappa-var-ser-jag-mina-egna-relationer]]
- **notes/rekrytering-av-it-personal-i-sundsvall-marknadsbild-och-frolista** <- [[notes/it-arbetsplatser-i-sundsvall-3-469-enligt-scb-yrkesregistret]], [[notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen]], [[notes/pappa-anstallningshistorik-i-cv-format-fix-av-kompetensordboken]], [[notes/pappa-uppdragsmatchning-delpoang-forklarande-ranking]], [[notes/utbyggnad-av-pappas-personlista-bolagsregister-konferensbios-gav-140-namngivna-personer]]
- **notes/rekryteringspipeline-board-per-person-sourcing-som-attribut** <- [[notes/relationer-mellan-kandidater-i-pappa-symmetriska-par-med-styrka-1-10]]
- **notes/relationer-kan-nu-laggas-till-fran-oversikten-dialogen-har-par-lage** <- [[notes/relationsoversikt-i-pappa-var-ser-jag-mina-egna-relationer]]
- **notes/relationer-mellan-kandidater-i-pappa-symmetriska-par-med-styrka-1-10** <- [[notes/detaljvy-for-kandidat-i-pappa-gemensamma-vagar-mellan-kandidater]]
- **notes/relationsgrafen-anvander-fast-michael-id-utan-person-kind** <- [[notes/michael-brundin-blir-en-personnod-i-relationsgrafen]]
- **notes/relationsoversikt-i-pappa-var-ser-jag-mina-egna-relationer** <- [[notes/personer-kan-laggas-in-manuellt-i-pappa]], [[notes/redigerare-och-agarskap-for-relationer-i-pappa]], [[notes/relationer-kan-nu-laggas-till-fran-oversikten-dialogen-har-par-lage]]
- **notes/repo-template-ci-gitops-flodet-fixat-bygger-3-images-ghcr-tagging-pr-overlays-per-pr** <- [[notes/pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres]]
- **notes/repo-template-ny-rust-api-src-rust-med-axum-utoipa-scalar** <- [[notes/vertikal-kalender-med-ikonstod-i-plan]]
- **notes/repo-template-rust-api-kopplad-till-kubernetes-manifests-tilt** <- [[notes/tilt-filbevakning-kan-do-starta-om-tilt-up]]
- **notes/rescue-wipe-pa-rock-maste-tomma-hela-systemdisken-wipe-system-esp-filen-ar-det-som-fangar-booten** <- [[notes/pxe-bootloop-initramfs-startar-om-for-att-boota-installerad-talos-pa-emmc]]
- **notes/rk3399-s-bootrom-soker-spi-emmc-sd-usb-hardkodat** <- [[entities/pxe-server]], [[notes/pa-rock-plattorna-finns-ingen-bootloader-i-hardvaran]], [[notes/radxa-bradena-behover-bootmedium-men-talos-och-bootloader-maste-ligga-pa-skilda-medium]], [[notes/u-boot-2026-01-rock-4c-saknar-bootfile-boot-script-dhcp-natverksboot-dor-direkt]], [[notes/u-boot-kors-en-gang-boot-targets-styr-steg-2-inte-steg-1]]
- **notes/rock-4c-4se-talos-kan-aldrig-installeras-till-disk-u-boot-saknar-efi-setvariablert** <- [[notes/u-boot-for-rock-4c-built-from-mainline-with-pxe-first]]
- **notes/rock-4c-4se-utan-batteribackup-rtc-star-pa-2017-tls-mot-factory-talos-dev-misslyckas** <- [[notes/pxe-genererade-nodconfigs-var-stale-efter-talos-v1-14-uppgraderingen]]
- **notes/rock-4c-och-4se-har-ingen-spi-flash-i-fabrik-u-boot-ligger-pa-microsd-emmc** <- [[entities/pxe-server]], [[notes/korrigering-rock4c-plus-bin-ar-ett-microsd-emmc-image-inte-en-spi-image]], [[notes/pxe-forst-bootloader-for-rock-4c-byggd-som-patchad-u-boot-image]], [[notes/worker-1-rock-4c-nar-pxe-via-u-boot-boot-targets-men-pxe-ligger-sist]]
- **notes/rock-pi-4-rk3399-defconfig-bygger-en-rock-pi-4a-for-4c-ska-rock-4c-plus-rk3399-defconfig-anvandas** <- [[entities/pxe-server]], [[notes/mellanslag-i-radxa-rock-4c-saboterar-byggen-losning-bind-mount]]
- **notes/root-dir-funktionen-stoppar-no-project-root-found-meddelandet** <- [[notes/rust-utvecklingsmiljo-pa-workstationen]]
- **notes/rust-analyzer-standalone-lage-kraschar-pa-leetcode-nvim-filer** <- [[notes/root-dir-funktionen-stoppar-no-project-root-found-meddelandet]]
- **notes/rust-analyzer-tysta-health-status-notisen-istallet-for-att-stanga-av-standalone** <- [[notes/rust-analyzer-standalone-lage-kraschar-pa-leetcode-nvim-filer]]
- **notes/rust-learning-repot-lades-in-som-submodule-mot-egen-fork** <- [[notes/rust-utvecklingsmiljo-pa-workstationen]]
- **notes/rust-repo-learn-to-code-with-rust-som-submodule** <- [[notes/rust-learning-repot-lades-in-som-submodule-mot-egen-fork]]
- **notes/rust-utvecklingsmiljo-pa-workstationen** <- [[notes/rust-repo-learn-to-code-with-rust-som-submodule]]
- **notes/sa-dampas-bada-rust-analyzer-popupen-no-project-root-found** <- [[notes/rust-analyzer-tysta-health-status-notisen-istallet-for-att-stanga-av-standalone]]
- **notes/sales-soulmate-se-appen-kan-inte-logga-in-oauth-clienten-saknas-i-soulmate-tenanten** <- [[notes/imap-inloggning-for-michael-brundin-soulmate-se-kraver-admin-consent]]
- **notes/session-livscykel** <- [[notes/accuracy-fel-omstart]], [[notes/arkitektur]], [[notes/sprint-timer-saknas]]
- **notes/simon-kvm-cli-kommando-for-att-byta-hdmi-kvm-port-fran-pxe-vard** <- [[notes/kvm-styrning-fran-pxe-pi-n-16-port-hdmi-kvm]]
- **notes/simon-pxe-flash-clean-code-omgang-2-fokuserade-funktioner** <- [[entities/pxe-server]]
- **notes/simon-pxe-flash-flashar-pxe-forst-u-boot-till-sd-kort-i-pxe-servern** <- [[entities/pxe-server]]
- **notes/simonbrundin-gmail-com-inkorgen-kan-lasas-via-imap-begbot-losenordet** <- [[notes/imap-inloggning-for-michael-brundin-soulmate-se-kraver-admin-consent]]
- **notes/skogsduvan-elementet-styrs-av-minsta-temperaturen-fail-open** <- [[notes/skogsduvan-temperaturkaskad-i-tre-steg-styrd-av-minsta-temperaturen]]
- **notes/skogsduvan-fyra-shelly-enheter-svarar-inte-pa-natet** <- [[notes/skogsduvans-ha-temperaturlogik-ombyggd-fyra-doda-sensorer]]
- **notes/skogsduvan-larm-nar-sensorer-shelly-enheter-faller-ur-online** <- [[notes/skogsduvan-temperaturkaskad-i-tre-steg-styrd-av-minsta-temperaturen]]
- **notes/skogsduvan-ssh-via-duckdns-fungerar-port-22-forwardad-2026-11-02** <- [[notes/skogsduvan-temperaturkaskad-i-tre-steg-styrd-av-minsta-temperaturen]]
- **notes/skogsduvan-temperaturkaskad-i-tre-steg-styrd-av-minsta-temperaturen** <- [[notes/ha-automationer-behover-en-startv-g-for-att-aterstalla-laget-och-tre-syntaxfallor]], [[notes/skogsduvan-elementet-styrs-av-minsta-temperaturen-fail-open]], [[notes/skogsduvan-larm-nar-sensorer-shelly-enheter-faller-ur-online]], [[notes/skogsduvans-ha-temperaturlogik-ombyggd-fyra-doda-sensorer]]
- **notes/skogsduvans-ha-temperaturlogik-ombyggd-fyra-doda-sensorer** <- [[notes/skogsduvan-fyra-shelly-enheter-svarar-inte-pa-natet]], [[notes/skogsduvan-temperaturkaskad-i-tre-steg-styrd-av-minsta-temperaturen]], [[notes/skogsduvans-ha-utsatts-for-automatiserade-skannningar-fran-internet]]
- **notes/skogsduvans-ha-utsatts-for-automatiserade-skannningar-fran-internet** <- [[notes/skogsduvan-temperaturkaskad-i-tre-steg-styrd-av-minsta-temperaturen]]
- **notes/sortering-av-grundmal-pa-goals-med-piltangenterna** <- [[notes/enter-pa-goals-skapar-nytt-grundmal-istallet-for-att-oppna-malet]], [[notes/prioriteringsordningen-pa-now-sparades-inte-bulk-update-och-spar-ko]]
- **notes/sortering-av-undermal-med-piltangenterna-var-trasig-i-plan** <- [[notes/barnordning-foljer-beroenden-i-malsidan]], [[notes/malgraph-navigering-mellan-malsidor-gar-fran-sekunder-till-millisekunder]], [[notes/sortering-av-grundmal-pa-goals-med-piltangenterna]]
- **notes/sprint-timer-saknas** <- [[notes/session-livscykel]]
- **notes/standardkalender-i-plan** <- [[notes/flera-kalendrar-med-egna-farger-i-plan]], [[notes/handelsen-har-ingen-egen-farg]]
- **notes/started-finished-kunde-inte-rensas-i-plan-null-ignorerades-i-go-api-t** <- [[notes/authflytten-i-plan-sessionen-ligger-nu-i-go-api-t-inte-i-nuxt]]
- **notes/stk-dev-miljon-kor-mot-neon-istallet-for-cnpg** <- [[notes/pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres]]
- **notes/stk-matchi-synk-sjalvaterstallande-synk-och-var-database-url-satts** <- [[notes/synk-hallbokningar-kommunens-actor-smartbook-matchi]]
- **notes/stray-wiki-i-typetrainer-repot-ar-inte-kallan-for-wikin** <- [[notes/kunskaps-wiki]]
- **notes/synk-hallbokningar-kommunens-actor-smartbook-matchi** <- [[notes/krypterade-secrets-i-klustret-ogiltiga-inloggningar-och-keycloak-brute-force]], [[notes/matchi-bokning-slot-id-andras-nar-bokningen-skapas]], [[notes/matchi-keycloak-session-code-maste-fangas-med-generell-teckenklass]], [[notes/stk-matchi-synk-sjalvaterstallande-synk-och-var-database-url-satts]]
- **notes/talos-etcd-en-gammal-ephemeral-kopia-pa-controlplanet-ger-en-trasig-raft-log** <- [[notes/controlplane-recovery-via-pxe-ar-inte-konfigurerad]]
- **notes/talos-hostname-auto-i-en-generisk-config-byter-nodens-namn-och-kastar-ut-den-ur-klustret** <- [[notes/talos-v1-14-machineconfig-specen-ar-en-quoted-string-och-flera-dokument]]
- **notes/talos-image-cache-kraver-ephemeral-cirkulart-beroende-med-state-meta-pa-en-farsk-metal-nod** <- [[notes/gotcha-pxe-forst-gor-noden-till-en-pxe-nod-maintenance-mode]]
- **notes/talos-installen-nollstaller-de-forsta-11-mb-pa-microsd-loadern-maste-aterstallas** <- [[notes/radxa-bradena-behover-bootmedium-men-talos-och-bootloader-maste-ligga-pa-skilda-medium]]
- **notes/talos-noden-kan-hamna-i-maintenance-mode-om-state-volymen-hittas-efter-config-fasen** <- [[notes/gotcha-pxe-forst-gor-noden-till-en-pxe-nod-maintenance-mode]]
- **notes/talos-v1-14-machineconfig-specen-ar-en-quoted-string-och-flera-dokument** <- [[entities/talos-linux]], [[notes/pxe-genererade-nodconfigs-var-stale-efter-talos-v1-14-uppgraderingen]]
- **notes/talos-v1-14-utan-unattendedinstallconfig-oinstallerad-metal-nod-bootar-om-i-en-loop** <- [[notes/pxe-bootloop-initramfs-startar-om-for-att-boota-installerad-talos-pa-emmc]]
- **notes/tidsatgang-pa-mal-i-plan-estimated-minutes** <- [[notes/undo-stapel-i-plan-z-angrar-senaste-malhandling]]
- **notes/tilt-filbevakning-kan-do-starta-om-tilt-up** <- [[notes/pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres]]
- **notes/transienta-502-fra-go-api-t-nar-poddar-startar-om-i-dev-klustret** <- [[notes/pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres]]
- **notes/tre-satt-att-pxe-boota-rock-4c-och-varfor-dhcp-metoden-valdes** <- [[entities/pxe-server]]
- **notes/twingate-anslutaren-for-10-10-10-0-24-var-nere-recovery-floden-kan-inte-koras-fran-arbetsstationen** <- [[notes/skogsduvan-ssh-via-duckdns-fungerar-port-22-forwardad-2026-11-02]]
- **notes/twingate-ressurserna-forsvann-pa-arbetsstationen-2026-11-04-hemnatet-10-1-1-1-och-datacenter-10-10-10-0-24-oppnarbara** <- [[notes/twingate-anslutaren-for-10-10-10-0-24-var-nere-recovery-floden-kan-inte-koras-fran-arbetsstationen]]
- **notes/u-boot-2026-01-rock-4c-saknar-bootfile-boot-script-dhcp-natverksboot-dor-direkt** <- [[entities/pxe-server]], [[notes/korrigering-rock4c-plus-bin-ar-ett-microsd-emmc-image-inte-en-spi-image]], [[notes/pa-rock-plattorna-finns-ingen-bootloader-i-hardvaran]], [[notes/rk3399-s-bootrom-soker-spi-emmc-sd-usb-hardkodat]], [[notes/rock-4c-och-4se-har-ingen-spi-flash-i-fabrik-u-boot-ligger-pa-microsd-emmc]], [[notes/u-boot-kors-en-gang-boot-targets-styr-steg-2-inte-steg-1]]
- **notes/u-boot-for-rock-4c-built-from-mainline-with-pxe-first** <- [[entities/pxe-server]], [[notes/korrigering-rock4c-plus-bin-ar-ett-microsd-emmc-image-inte-en-spi-image]], [[notes/pxe-boot-script-dhcp-kraver-mkimage-uimage-inte-ren-hush-text]], [[notes/pxe-kedjan-for-worker-1-verifierad-serverside-pxe-metoden-failar-snabbt-utan-pxelinux-cfg]], [[notes/rock-4c-och-4se-har-ingen-spi-flash-i-fabrik-u-boot-ligger-pa-microsd-emmc]], [[notes/rock-pi-4-rk3399-defconfig-bygger-en-rock-pi-4a-for-4c-ska-rock-4c-plus-rk3399-defconfig-anvandas]], [[notes/tre-satt-att-pxe-boota-rock-4c-och-varfor-dhcp-metoden-valdes]], [[notes/worker-1-kortet-flashat-med-ny-mainline-build-tf-a-fran-repot]], [[notes/worker-1-rock-4c-tyst-mainline-u-boot-byggd-men-boarden-laser-inte-kortet]]
- **notes/u-boot-hamtar-alltid-boot-script-dhcp-per-nod-config-via-ipaddr** <- [[entities/pxe-server]]
- **notes/u-boot-kors-en-gang-boot-targets-styr-steg-2-inte-steg-1** <- [[entities/pxe-server]], [[notes/gotcha-pxe-forst-gor-noden-till-en-pxe-nod-maintenance-mode]], [[notes/pa-rock-plattorna-finns-ingen-bootloader-i-hardvaran]], [[notes/pxe-forst-bootloader-for-rock-4c-byggd-som-patchad-u-boot-image]], [[notes/pxe-kedjan-dhcp-boot-pekar-pa-textskript-matchbox-efi-ar-uefi-vagen]], [[notes/rk3399-s-bootrom-soker-spi-emmc-sd-usb-hardkodat]], [[notes/u-boot-2026-01-rock-4c-saknar-bootfile-boot-script-dhcp-natverksboot-dor-direkt]]
- **notes/u7-lite-94-2a-6f-7a-4b-33-adopterar-inte-saknar-management-ip-och-nattverkssegment-pa-switchen** <- [[notes/picokvm-luckfox-startar-inte-i-labbet-lite-saknar-emmc-ingen-poe]]
- **notes/undo-stapel-i-plan-z-angrar-senaste-malhandling** <- [[notes/tidsatgang-pa-mal-i-plan-estimated-minutes]]
- **notes/unifi-gatewayen-har-ingen-publik-ipv6-cgnat-via-starlink** <- [[notes/ipv6-ingress-till-envoy-designlage-2026-11-15]]
- **notes/usb-disken-acer-wd-320gb-status-och-longhorn-bedomning** <- [[concepts/longhorn]], [[entities/talos-linux]]
- **notes/utbyggnad-av-pappas-personlista-bolagsregister-konferensbios-gav-140-namngivna-personer** <- [[notes/pappa-anstallningshistorik-i-cv-format-fix-av-kompetensordboken]], [[notes/rekrytering-av-it-personal-i-sundsvall-marknadsbild-och-frolista]]
- **notes/uttunnade-serier-vags-sa-att-antalen-blir-exakta** <- [[notes/en-uttunnad-forekomst-maste-vara-synlig-aven-om-den-ligger-utanfor-fonstret]]
- **notes/utvecklingsflode** <- [[notes/arkitektur]], [[notes/konventioner]], [[notes/kunskaps-wiki]], [[notes/median-i-stallet-for-snitt-for-wpm]], [[notes/wordlist-kallor]], [[summaries/llm-wiki-karpathy]]
- **notes/valfri-fargkod-pa-kalendrar** <- [[notes/flera-kalendrar-med-egna-farger-i-plan]]
- **notes/verifierade-policy-hal-i-tool-lagret-kommandon-ar-blocklist-inte-allowlist** <- [[notes/artikeln-om-agentic-software-factory-vad-som-ar-relevant-for-agentsystemet]]
- **notes/vertikal-kalender-med-ikonstod-i-plan** <- [[notes/flera-kalendrar-med-egna-farger-i-plan]], [[notes/kalenderns-detaljnivaer-fem-nivaer-och-lane-packning]], [[summaries/kalenderns-detaljnivaer-och-lane-packning]]
- **notes/vertikal-kalender-med-kontinuerlig-zoom** <- [[notes/kalenderns-detaljnivaer-fem-nivaer-och-lane-packning]], [[notes/kalenderns-nivavaxling-animeras-crossfade-pop-in-och-hysteresis]], [[notes/parallella-agent-sessioner-i-samma-repo-och-databas]], [[notes/vertikal-kalender-med-ikonstod-i-plan]], [[summaries/kalenderns-detaljnivaer-och-lane-packning]]
- **notes/wordlist-kallor** <- [[notes/bigram-algoritmen]]
- **notes/worker-1-kortet-flashat-med-ny-mainline-build-tf-a-fran-repot** <- [[entities/pxe-server]], [[notes/pxe-boot-script-dhcp-kraver-mkimage-uimage-inte-ren-hush-text]], [[notes/pxe-kedjan-for-worker-1-verifierad-serverside-pxe-metoden-failar-snabbt-utan-pxelinux-cfg]], [[notes/pxe-mainline-u-boots-bootefi-kraver-addr-size-for-image-och-initrd]], [[notes/rock-pi-4-rk3399-defconfig-bygger-en-rock-pi-4a-for-4c-ska-rock-4c-plus-rk3399-defconfig-anvandas]], [[notes/simon-pxe-flash-flashar-pxe-forst-u-boot-till-sd-kort-i-pxe-servern]], [[notes/tre-satt-att-pxe-boota-rock-4c-och-varfor-dhcp-metoden-valdes]], [[notes/u-boot-for-rock-4c-built-from-mainline-with-pxe-first]], [[notes/u-boot-hamtar-alltid-boot-script-dhcp-per-nod-config-via-ipaddr]], [[notes/worker-1-rock-4c-tyst-mainline-u-boot-byggd-men-boarden-laser-inte-kortet]], [[notes/worker-1-svarar-inte-i-natet-felsokningsstatus]]
- **notes/worker-1-och-worker-2-ar-doda-pa-link-niva-pxe-kedjan-ar-inte-blockerande** <- [[entities/pxe-server]], [[notes/worker-1-rock-4c-tyst-mainline-u-boot-byggd-men-boarden-laser-inte-kortet]], [[notes/worker-1-startar-om-med-strom-men-skickar-noll-paket-fysisk-omstart-bevisar-samma-fel]], [[notes/worker-1-svarar-inte-ens-med-en-full-poe-cykel-noll-paket-i-bade-riktningar]], [[notes/worker-1-svarar-inte-i-natet-felsokningsstatus]], [[notes/worker-2-utan-strom-port-22-up-false-poe-power-0]]
- **notes/worker-1-rock-4c-nar-pxe-via-u-boot-boot-targets-men-pxe-ligger-sist** <- [[entities/pxe-server]], [[notes/gotcha-pxe-forst-gor-noden-till-en-pxe-nod-maintenance-mode]], [[notes/pxe-forst-bootloader-for-rock-4c-byggd-som-patchad-u-boot-image]], [[notes/pxe-mainline-u-boots-bootefi-kraver-addr-size-for-image-och-initrd]], [[notes/rock-4c-och-4se-har-ingen-spi-flash-i-fabrik-u-boot-ligger-pa-microsd-emmc]], [[notes/simon-pxe-flash-flashar-pxe-forst-u-boot-till-sd-kort-i-pxe-servern]], [[notes/u-boot-hamtar-alltid-boot-script-dhcp-per-nod-config-via-ipaddr]]
- **notes/worker-1-rock-4c-tyst-mainline-u-boot-byggd-men-boarden-laser-inte-kortet** <- [[entities/pxe-server]], [[notes/mellanslag-i-radxa-rock-4c-saboterar-byggen-losning-bind-mount]], [[notes/pxe-boot-script-dhcp-kraver-mkimage-uimage-inte-ren-hush-text]], [[notes/pxe-kedjan-for-worker-1-verifierad-serverside-pxe-metoden-failar-snabbt-utan-pxelinux-cfg]], [[notes/pxe-mainline-u-boots-bootefi-kraver-addr-size-for-image-och-initrd]], [[notes/simon-pxe-flash-flashar-pxe-forst-u-boot-till-sd-kort-i-pxe-servern]], [[notes/tre-satt-att-pxe-boota-rock-4c-och-varfor-dhcp-metoden-valdes]], [[notes/u-boot-for-rock-4c-built-from-mainline-with-pxe-first]], [[notes/u-boot-hamtar-alltid-boot-script-dhcp-per-nod-config-via-ipaddr]], [[notes/worker-1-kortet-flashat-med-ny-mainline-build-tf-a-fran-repot]], [[notes/worker-1-och-worker-2-ar-doda-pa-link-niva-pxe-kedjan-ar-inte-blockerande]], [[notes/worker-1-rock-4c-nar-pxe-via-u-boot-boot-targets-men-pxe-ligger-sist]], [[notes/worker-1-svarar-inte-i-natet-felsokningsstatus]]
- **notes/worker-1-startar-om-med-strom-men-skickar-noll-paket-fysisk-omstart-bevisar-samma-fel** <- [[notes/workers-bootar-inte-via-pxe-trots-fungerande-pxe-server-unifi-poe-styrning-via-rest-device]]
- **notes/worker-1-svarar-inte-ens-med-en-full-poe-cykel-noll-paket-i-bade-riktningar** <- [[notes/worker-1-startar-om-med-strom-men-skickar-noll-paket-fysisk-omstart-bevisar-samma-fel]]
- **notes/worker-1-svarar-inte-i-natet-felsokningsstatus** <- [[entities/pxe-server]], [[notes/worker-1-och-worker-2-ar-doda-pa-link-niva-pxe-kedjan-ar-inte-blockerande]]
- **notes/worker-2-utan-strom-port-22-up-false-poe-power-0** <- [[entities/pxe-server]], [[notes/worker-1-och-worker-2-ar-doda-pa-link-niva-pxe-kedjan-ar-inte-blockerande]]
- **notes/worker-7-s-kubeletcertifikat-signerades-av-en-gammal-kubernetes-ca-som-finns-kvar-i-git-historiken** <- [[notes/worker-1-svarar-inte-ens-med-en-full-poe-cykel-noll-paket-i-bade-riktningar]], [[notes/worker-8-korrupt-xfs-pa-ephemeral-partitionen-gjorde-att-kubelet-aldrig-startade-fix-med-mkfs-xfs]]
- **notes/worker-8-korrupt-xfs-pa-ephemeral-partitionen-gjorde-att-kubelet-aldrig-startade-fix-med-mkfs-xfs** <- [[notes/cluster1-stale-kubeconfig-ca-och-noll-workers-diagnos-2026-11-11]], [[notes/worker-7-s-kubeletcertifikat-signerades-av-en-gammal-kubernetes-ca-som-finns-kvar-i-git-historiken]]
- **notes/workers-bootar-inte-via-pxe-trots-fungerande-pxe-server-unifi-poe-styrning-via-rest-device** <- [[entities/pxe-server]], [[notes/cluster1-stale-kubeconfig-ca-och-noll-workers-diagnos-2026-11-11]], [[notes/ipv6-ingress-till-envoy-designlage-2026-11-15]]
- **summaries/16x1-hdmi-kvm-rs232-lan-protokoll** <- [[entities/hdmi-kvm-switch-16-port-labbet]], [[notes/kvm-styrning-fran-pxe-pi-n-16-port-hdmi-kvm]]
- **summaries/2026-11-03-learn-kunskapsbasen-kan-lagras-i-postgresql** <- [[notes/learn-kunskapsbasen-kan-lagras-i-postgresql]]
- **summaries/flux-components** <- [[concepts/flux-operator]], [[entities/flux-gitops]], [[entities/kubernetes]]
- **summaries/infrastructure-flux** <- [[concepts/flux-operator]], [[concepts/gitops]]
- **summaries/kalenderns-detaljnivaer-och-lane-packning** <- [[notes/kalenderns-detaljnivaer-fem-nivaer-och-lane-packning]]
- **summaries/llm-wiki-karpathy** <- [[notes/kunskaps-wiki]], [[notes/utvecklingsflode]]
- **summaries/longhorn-disk-guide** <- [[concepts/longhorn]], [[entities/talos-linux]], [[notes/talos-v1-14-machineconfig-specen-ar-en-quoted-string-och-flera-dokument]]
