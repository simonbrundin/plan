# Cluster

- [x] Uppdatera Agents.md med information om hur min infrastucture fungerar
- [ ] ´simon cluster health´ grönt
  - [x] Få igång worker-2
  - [ ] Få alla Flux Komponenter att fungera
  - [ ] Föra över de sista komponenterna från ArgoCD till Flux
  - [x] Få igång worker-9
  - [x] Alla ArgoCD-poddar körs
  - [x] skapa ett simon kommando som skapar en secret.yaml som är krypterad.
        antingen med ett genegererat lösenord eller med ett jag väljer tex en
        api nyckel
  - [ ] Sätt upp webhook för flux
  - [ ] Köra Config Sync i klustret så det bara går att göra förändringar via
        Git.
        [🚀 KRM-Native GitOps: Yes — Without Flux, No. (FluxCD or Nothing.)](https://www.linkedin.com/pulse/krm-native-gitops-yes-without-flux-fluxcd-nothing-mialon-wsmue/)
- [x] Just nu känns det som upgrade körs varje gång jag kör simon talos update
      config. så ska det ju inte vara.
- [ ] Lägga till en Kyverno-policy som gör det omöjligt att radera
      longhorn-system namespacet
  - [ ] Installera Kyverno
- [x] Köra kubescape eller liknande för att se säkerhetsproblem i mitt kluster
- [ ] Jag vill kunna se en visuell graf av min databas schema som jag kan se i
      både prod och dev med tex tilt
  - [ ] Chardb
- [x] Longhorn rapporterar grönt
  - [x] 2 - Inte påslagen, troligtvis kör den igång från fel disk
  - [x] 3 - Disken full
  - [x] 5 - Disk 2 inte inkopplad
  - [x] 6 - Disken full
  - [x] 7 - Disken full
  - [x] 9 - Systemdisk måste sättas i patch
  - [x] ´simon kubernetes longhorn test´ för alla noder och se till att allt
    - [x] Systemdisk blir grönt när konfiguerad serial är samma som installerad
          är grönt
    - [x] Extensions visar iscsi-tools och inte schematic
- [ ] Uppdatera istallationsguide för en ny nod eller disk så rätt serial läggs
      till i patchen
- [ ] Ändra installationsdisk för worker-2
- [x] Få igång Vault backup
- [ ] Koppla in bilbatteri
- [x] Uppgradera Kubernetes
- [ ] Installera Talos på datorer
  - [x] Mac Mini
  - [ ] HP
- [ ] Dot AI
- [ ] Teleport fungerar
- [ ] Vault fungerar
- [ ] auth.simonbrundin.com fungerar
- [ ] Installera
      [Kubernetes Dashboard](https://kubernetes.io/docs/tasks/access-application-cluster/web-ui-dashboard/)
- [ ] Installera Renovate i klustret så jag får PRs om uppdateringar

---

---

## Back on Kubernetes-track

- [x] Kubelet CSR Approver fungerar
- [x] ArgoCD fungerar

### Longhord fungerar

- [x] Få worker-5 stabil i csi-plugin och manager
- [x] Förstå om det bara är dessa som fungerar egentligen. kubectl get
      volumes.longhorn.io -n longhorn-system volumes.longhorn.io -n
      longhorn-system
- [x] Förstå om det bara är dessa som fungerar egentligen. kubectl get
      volumes.longhorn.io -n longhorn-system volumes.longhorn.io -n
      longhorn-system
