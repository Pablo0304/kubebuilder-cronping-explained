# CronPing Operator (Kubebuilder) - Guía Completa

## Propósito

Este repositorio contiene un operador de Kubernetes construido con Kubebuilder.
El objetivo es definir un recurso personalizado (CRD) para programar pings HTTP
y generar CronJobs de Kubernetes que ejecuten esos pings en el cluster.

Este proyecto sirve como ejemplo realista de buenas prácticas:

- Separación entre configuración (PingTarget) y planificación (CronPing).
- Controlador que sincroniza el estado deseado con el estado real.
- Uso de CRDs, RBAC, kustomize y un ciclo de desarrollo local con kind.

## Conceptos clave (resumen rápido)

- CRD (CustomResourceDefinition): define un tipo nuevo en Kubernetes.
- CR (CustomResource): instancia concreta de ese tipo (tu YAML).
- Controller/Reconciler: código que observa CRs y crea/actualiza recursos reales.
- Group/Version/Kind (GVK): nombre canónico del tipo en Kubernetes.

## Arquitectura del ejemplo

Se usan dos CRDs en el mismo grupo `demo.my.domain`:

1. PingTarget (configuración reutilizable)
   - Define a dónde hacer ping (URL, headers, etc.)
2. CronPing (planificación)
   - Define cada cuánto hacer ping y a qué target apunta

El controlador de CronPing:

- Lee el CronPing y su PingTarget
- Crea un CronJob con un contenedor curl
- El CronJob ejecuta el ping en el horario indicado

## Requisitos

- Go instalado (versión recomendada por Kubebuilder)
- kubebuilder instalado
- Docker instalado (para kind)
- kind instalado
- kubectl instalado

## Estructura relevante del repositorio

- `api/v1alpha1/`
  - Tipos Go para los CRDs (CronPing y PingTarget)
- `internal/controller/`
  - Lógica de reconciliación
- `config/`
  - Manifiestos kustomize (CRDs, RBAC, manager)
- `config/samples/`
  - Ejemplos de CRs para pruebas

## Pasos para inicializar (orden recomendado)

1. Inicializar proyecto Kubebuilder:

   - `kubebuilder init --domain my.domain --repo my.domain/cronping-operator`

2. Crear CRD CronPing:

   - `kubebuilder create api --group demo --version v1alpha1 --kind CronPing`

3. Crear CRD PingTarget:

   - `kubebuilder create api --group demo --version v1alpha1 --kind PingTarget`

4. Editar la lógica del controller:

   - Archivo: `internal/controller/cronping_controller.go`
   - Objetivo: crear un CronJob a partir de CronPing + PingTarget

5. Generar manifiestos (CRDs/RBAC):

   - `make manifests`

6. Crear un cluster local con kind (en este caso yo ya lo había creado):

   - `kind create cluster --name kubebuilder`

7. Instalar CRDs en el cluster:

   - `make install`

8. Ejecutar el controlador local:

   - `make run`

9. Aplicar ejemplos (previamente hay que editar estos ".yaml"):

   - `kubectl apply -k config/samples`

## Ejecución local vs despliegue en el cluster

- `make run`: ejecuta el controlador en tu máquina. Es ideal para desarrollo local, pero se detiene si cierras la terminal.
- `make deploy`: instala el operador dentro del cluster (Deployment). Es el modo recomendado para entornos reales.

Flujo típico:

1. `make install` (instala CRDs)
2. `make deploy` (despliega el operador en el cluster)
3. `make undeploy` (elimina el operador del cluster)

## Definición de CRDs (explicación)

### CronPing

Archivo: `api/v1alpha1/cronping_types.go`

Campos principales:

- `spec.targetRef`: referencia al PingTarget (nombre)
- `spec.schedule`: expresión cron (ej: "_/2 _ \* \* \*")

### PingTarget

Archivo: `api/v1alpha1/pingtarget_types.go`

Campos principales:

- `spec.url`: URL a la que se hace ping
- `spec.method`: método HTTP (opcional)
- `spec.timeoutSeconds`: timeout (opcional)
- `spec.headers`: headers opcionales

## Ejemplos (samples)

Los ejemplos viven en `config/samples/` y se aplican con:

- `kubectl apply -k config/samples`

Ejemplo de PingTarget:

```yaml
apiVersion: demo.my.domain/v1alpha1
kind: PingTarget
metadata:
  name: google
  namespace: dev
spec:
  url: https://www.google.com
```

Ejemplo de CronPing:

```yaml
apiVersion: demo.my.domain/v1alpha1
kind: CronPing
metadata:
  name: ping-google
  namespace: dev
spec:
  targetRef: google
  schedule: "*/2 * * * *"
```

## Cómo comprobar que funciona

1. Ver CRs:

- `kubectl get pingtargets -n dev`
- `kubectl get cronpings -n dev`

2. Ver CronJob creado:

- `kubectl get cronjob -n dev`

3. Ver Jobs y Pods:

- `kubectl get jobs -n dev`
- `kubectl get pods -n dev`

4. Ver logs del pod:

- `kubectl logs -n dev -l job-name=ping-google`

## Buenas prácticas aplicadas

- Un operador por dominio funcional (ping/monitorización básica).
- CRDs relacionados dentro del mismo grupo.
- Separar configuración (PingTarget) de planificación (CronPing).
- Uso de RBAC mínimo necesario.
- Kustomize para aplicar recursos agrupados.

## Troubleshooting rápido

- "no matches for kind ...": falta instalar CRD -> ejecutar `make install`.
- "resource mapping not found": apiVersion incorrecta -> usar `demo.my.domain/v1alpha1`.
- No se ven recursos en `dev`: aplicar YAML con `metadata.namespace: dev`.

## Siguientes pasos recomendados

- Actualizar CronJob cuando cambie el spec.
- Escribir `status` en CronPing (último ping, ok/fail).
- Validaciones OpenAPI (URL requerida, cron válido).
- Webhooks opcionales para validación avanzada.
