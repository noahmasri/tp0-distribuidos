# Parte 1

## Ejercicio 1
En el archivo `docker-compose-dev.yaml` se encuentra el archivo compose original, con el cual se realizaron pruebas para ver si la estructura que iba a desarrollar era la apropiada. Luego, en el archivo `docker-compose-test.yaml` se encuentra un archivo generado con el script de bash. Si bien la estructura de estos difiere levemente por problemas de indentación con la librería usada, su ejecución funciona como fue pedido. 

Se eligió usar la librería `pyyaml` en lugar de usar strings con el formato hecho a mano ya que considero que es mucho más seguro ante errores humanos. El uso de diccionarios para establecer las estructuras que luego mediante la propia librería se genera la indentación, hace que el código sea mucho mas legible y por tanto, menos propenso a errores.

### Como ejecutarlo
Antes de ejecutar el script, deberán asegurarse de contar con la dependencia requerida, en este caso, pyyaml. Para esto, se debe ejecutar el comando:
```bash
pip install PyYAML
```

Para generar el compose, debe correr el comando:
```bash
./generar-compose.sh <nombre archivo output> <numero de clientes>
```
En caso de recibir el error `bash: ./generar-compose.sh: Permission denied`, corra `chmod +x generar-compose.sh` para otorgar permisos de ejecución al script.

Es recomendable correr el comando con **docker-compose-dev.yaml** como nombre del archivo de output, ya que el makefile corre este archivo. En caso de desear otro nombre, deberá modificar el makefile. Una vez generado el compose, para ejecutar el programa, deberá usar los targets del makefile previamente mencionado. Los targets se ejecutan mediante la invocación de:

* **make \<target\>**:
Los target imprescindibles para iniciar y detener el sistema son **docker-compose-up** y **docker-compose-down**, siendo los restantes targets de utilidad para el proceso de _debugging_ y _troubleshooting_.

Los targets disponibles son:
* **docker-compose-up**: Inicializa el ambiente de desarrollo (buildear docker images del servidor y cliente, inicializar la red a utilizar por docker, etc.) y arranca los containers de las aplicaciones que componen el proyecto.
* **docker-compose-down**: Realiza un `docker-compose stop` para detener los containers asociados al compose y luego realiza un `docker-compose down` para destruir todos los recursos asociados al proyecto que fueron inicializados. Se recomienda ejecutar este comando al finalizar cada ejecución para evitar que el disco de la máquina host se llene.
* **docker-compose-logs**: Permite ver los logs actuales del proyecto. Acompañar con `grep` para lograr ver mensajes de una aplicación específica dentro del compose.
* **docker-image**: Buildea las imágenes a ser utilizadas tanto en el servidor como en el cliente. Este target es utilizado por **docker-compose-up**, por lo cual se lo puede utilizar para testear nuevos cambios en las imágenes antes de arrancar el proyecto.
* **build**: Compila la aplicación cliente para ejecución en el _host_ en lugar de en docker. La compilación de esta forma es mucho más rápida pero requiere tener el entorno de Golang instalado en la máquina _host_.

## Ejercicio 2
Se genera el ejercicio 2 a partir de lo realizado para el ejercicio 1. En esta rama se encuentran modificados tanto los dos compose que se encontraban de antes, como el generator para agregar un volume a cada contenedor que mapee de su archivo de configuración externo (ubicado en el sistema de archivos del host) al interior del contenedor.

Para verificar que el archivo de configuración se encuentre correctamente montado se puede usar el comando:

```bash
docker inspect <nombre_del_contenedor> --format='{{json .Mounts}}'
```
donde se espera, por ejemplo para el servidor, el output:
```bash
[{"Type":"bind","Source":"$(CWD)/server/config.ini","Destination":"/config.ini","Mode":"rw","RW":true,"Propagation":"rprivate"}]
```
siendo CWD el directorio en donde se encuentra el archivo compose.

Para ejecutar el EchoServer siga los mismos pasos que para el **ej1**.

## Ejercicio 3
Para llevar a cabo lo pedido, se creó un script que crea un contenedor que envía un mensaje al servidor ("Hello echoer!") y se conecta a este mediante la red de docker para no deber exponer el puerto del servidor por fuera de la red de contenedores. Este contenedor creado es temporal, es decir, luego de terminar su ejecución es eliminado, y se conecta a la red que define el archivo `docker-compose-dev.yaml`, tp0_testing_net.

Este script se basa principalmente en esta porción de código, que es la que se encarga de enviar el mensaje al servidor, y de obtener su respuesta. Esto lo que hace es guardarse en una variable response únicamente el resultado de la ejecución del comando `netcat`, deshaciendose tanto de la salida de stdout como la de stderr de todos los comandos. Lo primero que hace es instalar `netcat`, para lo cual había 2 versiones, traditional y openbsd, por lo que se debió elegir una. Luego hace echo del mensaje a enviar, enviando esto al servidor en el puerto en que este se encuentra, estableciendo arbitrariamente un timeout de 1 segundo.

```bash
RESPONSE=$(docker run --rm --network tp0_testing_net --name netcat ubuntu:18.04 bash -c "
    apt-get update > /dev/null 2>&1 && \
    apt-get install -y netcat-traditional > /dev/null 2>&1 && \ 
    echo \"$TEST_MESSAGE\" | nc -w 1 "server" $PORT
" 2>&1)
```
En la variable response entonces se guarda únicamente la respuesta del servidor. Se redirige a la salida estándar la salida de error para posible debugging. De querer saber por qué está fallando, se puede hacer echo del mensaje de response.

### Como ejecutarlo
Para que esto funcione, se debe tener la red corriendo, y para esto debemos correr el comando:
```bash
make docker-compose-up
```
En caso de que esta no esté corriendo, se recibirá el output `action: test_echo_server | result: fail`.
Una vez hecho esto correr:
```bash
./validar-echo-server.sh
```
En caso de recibir el error `bash: ./validar-echo-server.sh: Permission denied`, corra `chmod +x validar-echo-server.sh` para otorgar permisos de ejecución al script, y luego vuelva a ejecutar el comando anterior.
Si todo se encuentra corriendo y se recibe por parte del servidor el mismo mensaje que se le envió, se espera que el output sea `action: test_echo_server | result: success`.

## Ejercicio 4
Para este ejercicio se pidió que ambos el servidor y el cliente, al recibir una señal SIGTERM realicen un graceful shutdown. Para llevar a cabo esto en el servidor, se agregó a este un signal handler, que funciona mediante eventos. Si el servidor recibe la señal, automáticamente va a comenzar a ejecutarse la función establecida como handler, en este caso el método `__shutdown_gracefully`. Esta función lo que hace es settear el atributo `should_stop` en true, y cerrar el socket para no aceptar más conexiones. Al regresar a donde estaba, probablemente reciba un error intentando acceder al socket, por lo que este atributo sirve para ignorar el error que podría darse, ya que no es un error del SO sino que fue cerrado apropósito. A su vez, hice que el servidor se guarde como atributo el socket del cliente con el que se está comunicando para también cerrar este.
Por otro lado, la forma de manejar el shutdown desde el cliente es levemente distinta, debido a como maneja go las excepciones. Para manejar las excepciones en este, se debe escuchar mediante un channel, que en mi caso se llama notifier o sigchan. Este canal es escuchado en una goroutine que se encarga de, o bien manejar la señal recibida, o bien salir cuando se cierra el canal de comunicación con el hilo principal. Dentro del hilo principal, al terminar su ejecución se cierra el canal "done" para que la goroutine que maneja la excepción sepa que terminó y salga. Para que el hilo principal, se usa una barrera (waitgroup en go) para que este espere a la goroutina. Para que el hilo principal se entere si saltó la señal, se usa el mismo channel done, y en cada iteración este se fija si terminó el timer o recibió algo por done para ver si iterar una vez más o finalizar.

### Como ejecutarlo
Para verificar el correcto manejo de las señales, se debe correr:
```bash
make docker-compose-up
```
para iniciar los contenedores, y luego ejecutar:
```bash
make docker-compose-stop
```
para enviar la señal a los contenedores sin eliminarlos. Una vez hecho esto, se pueden observar los logs corriendo:
```bash
make docker-compose-logs
```
para ver las distintas acciones que ejecuta cada proceso al recibir SIGTERM. 

En caso de querer ver la ejecución llegar a su fin, saltear el segundo paso y acceder a los logs. Una vez visto lo deseado en los logs, puede salir de estos con ctrl+c, y cortar la ejecución de los contenedores con:
```bash
make docker-compose-down
```

# Parte 2
## Ejercicio 5
Para este ejercicio se pidió crear un protocolo de comunicación que permita que se emule una agencia de lotería, enviando por cliente una única apuesta. Este protocolo debe contener información suficiente para que se puedan evitar los short writes y los short reads.
Antes de desarrollar el protocolo en sí, se realizó un pequeño análisis de los datos otorgados por la cátedra, donde se encontró que:
* La máxima longitud de un nombre es de 23 caracteres en todos los csvs, mientras que la mínima se encuentra en la agencia 2, y es de 6 caracteres. Además, que una porción muy poco significativa de entradas tienen un apellido con ese máximo (menos de un 0,2%)
* La máxima longitud de un apellido es de 10 caracteres, mientras que la mínima se encuentra en las agencias 1 y 3 y es de 6 caracteres. Además, que una porción pequeña de entradas tienen un apellido con ese máximo (menos de un 2%)

Y además, se corroboró que:
* Todos los documentos tienen 8 caracteres. Por como están estructurados, sabemos que el máximo valor que podrían tomar es 99.999.999 (aunque en los datasets sea 40000000 aprox), valor el cual, eliminando los '.', entra en un u32.
* Las fechas tienen formato "YYYY-MM-DD"
* Las apuestas son todas de 4 digitos, numéricas, y con un valor máximo de 9999

Teniendo en cuenta todas estas validaciones sobre los datos es que se desarrolló un protocolo que pretende maximizar la cantidad de apuestas que se pueden realizar por paquete o por segundo, disminuyendo la cantidad de bytes utilizados por cada una de estas. El protocolo entonces plantea que en un mensaje viajará en el siguiente orden la siguiente información:

* Agencia: 1 byte, ya que por enunciado hay únicamente 5. 
* Longitud del nombre: para esto, dado que el máximo es 23 caracteres, se asigna 1 byte, u8, que será suficiente para alojar el tamaño. Se permite que puedan aparecer nombres más largos, siempre y cuando no superen los 255 caracteres
* Nombre: cantidad de bytes que se corresponden con la longitud. En base al byte anterior, se sabrá cuántos se deberá leer para obtener el nombre entero. 
* Longitud del apellido: para esto, dado que el máximo es 10 caracteres, se asigna 1 byte. Se permite que puedan aparecer nombres más largos, siempre y cuando no superen los 255 caracteres.
* Apellido: cantidad de bytes que se corresponden con la longitud.
* Documento: campo de longitud fija de 4 bytes, un u32, que permite alojar números que llegan hasta 4294967295, por lo que 100 millones entran en este tipo de números.
* Fecha: campo de longitud fija de 10 bytes. Se incluyen los '-', ya que es práctico almacenarlos así, y se considera que se podría tardar más parseando y desparseando este campo.
* Numero de apuesta: campo de longitud fija de 2 bytes, un u16, que permite alojar números que llegan hasta 65536, por lo que cualquier numero hasta el 9999 entran en este tipo de números.

Finalmente, se definen para el servidor 3 códigos de respuesta:
* Ok (0): si se pudo almacenar toda la información correctamente
* Error (1): si no se pudo almacenar la información, o no se la pudo leer por completo
* Abort (2): si el servidor recibió un sigterm o debió finalizar abruptamente por alguna señal

### Como ejecutarlo
Antes de ejecutarlo, debemos asegurarnos de tener las variables de entorno de la apuesta correctamente configuradas. Para esto, se debe encontrar en el cliente, dentro de la seccion environment del docker-compose.yaml algo similar a:
```yaml
environment:
    - CLI_ID=2
    - CLI_LOG_LEVEL=DEBUG
    - BET_NAME=<name of betmaker>
    - BET_SURNAME=<name of betmaker>
    - BET_ID=<Id number of betmaker>
    - BET_BIRTHDATE=<birthdate of betmaker in iso format (yyyy-mm-dd)>
    - BET_NUMBER=<number bet corresponds to>
```
Tal como en los anteriores ejercicios, debemos ejecutarlo corriendo los distintos targets del makefile. En este caso correríamos, para iniciar el servidor y los clientes:
```bash
make docker-compose-up
```
Para ver los logs de las apuestas:
```bash
make docker-compose-logs
```
Para finalizar la ejecución de todos los contenedores:
```bash
make docker-compose-down
```

## Ejercicio 6
Para este ejercicio, se pide leer apuestas de un csv, y enviarlas de a tandas al servidor. Para este, en primer lugar se pide definir la máxima cantidad de apuestas que se enviarán en un lote, si este lote no puede superar los 8kB de data. Dado el protocolo definido, sabemos que se envian por cada apuesta:
* long nombre: 1 byte por apuesta
* nombre: hasta 255 bytes por apuesta
* long apellido: 1 byte por apuesta
* apellido: hasta 255 bytes por apuesta
* documento: 4 bytes por apuesta
* fecha: 10 bytes por apuesta
* numero apuesta: 2 bytes por apuesta

Lo cual nos deja un máximo de 528 bytes por apuesta. En base a la investigación previa del set de datos, sabemos que el máximo en realidad es con 23 caracteres para el nombre, y 10 para el apellido, lo que nos da entonces en máximo de 51 bytes por apuesta. Con este número, tenemos entonces que la máxima cantidad de apuestas que entran en un batch son 156, ocupando 7956 bytes en el peor de los casos, y dejando la cantidad suficiente de bytes para agregar el número de agencia, y un campo nuevo que indique la cantidad de apuestas en el batch. 

### Lectura del CSV
En lugar de hacer una lectura por línea, para la lectura de los batches se optó por una solución un tanto más compleja, pero que ahorraba numerosas operaciones I/O. Se eligió arbitrariamente leer del archivo de a 4096 bytes, procesarlos separandolos por línea y generando apuestas con estas líneas. Para crear un batch luego se tiene que leer 1 o 2 veces de esta forma para llegar a los 156 del batch. Si en las dos lecturas en conjunto me paso del batchMax, entonces almaceno en mi estructura `BetGetter` las entradas sobrantes para la siguiente iteración.

### Modificación al protocolo
Siguiendo con la forma de enviar las apuestas anteriormente, se realizó una leve modificación al protocolo. Ahora se envía una única vez por batch el número de agencia en conjunto con la cantidad de apuestas en el batch. La forma de serializar y enviar cada apuesta sigue siendo la misma.
Para no saturar al servidor ni el buffer del TCP stream, se decidió ir enviando de a 1 batch, y en cuanto el servidor responde con la confirmación de lectura, entonces se manda el siguiente batch.

### Procesamiento en el servidor
Del lado del servidor, para recibir un batch lo que se hace es leer de a 1024 bytes del stream, e ir procesándolo de a tantas. Con esos 1024, se va avanzando y fraccionando los bytes pendientes de leer. Una vez que el querer parsear una apuesta de ese slice tira error cuando todavía tengo apuestas pendientes de leer, entonces volvemos a pedir 1024 bytes del stream, y así sucesivamente hasta alcanzar una cantidad de `batchSize` de apuestas.

### Cómo ejecutarlo
Como prerrequisitos, se pide que si no se tiene un archivo compose, se puede crear uno ejecutando el comando:

```bash
./generar-compose.sh docker-compose-dev.yaml <numero de clientes>
```
Además, se debe asegurar de contar con un archivo de data con la siguiente ruta:
```bash
/.data/agency-<id cliente>.csv
```
Teniendo la carpeta de data en la ruta al repositorio, y siendo el id del cliente el que se puede hallar en el compose como environment: CLI_ID.
Para ejecutarlo luego de estos requisitos se debe seguir los mismos pasos que en el ejercicio 5.

## Ejercicio 7
Para llevar a cabo lo pedido en este ejercicio, se debió agregar códigos de mensaje a los paquetes enviados por el cliente, y nuevos códigos de respuesta al servidor. Ahora el cliente cuenta con 3 mensajes:
* BET: para enviar un batch de apuesta
* END_BETTING: para avisar que las apuestas con las que él contaba ya las mandó, por lo que no va a estar enviando nada más
* REQUEST_WINNERS: para pedir que le envien los documentos de los DNIs ganadores de su agencia.
Para que se puedan enviar los resultados, tal como dice la consigna se necesita que 5 agencias hayan avisado que han finalizado el envío. Es por esto que las agencias se quedan en un loop (con cantidad de iteraciones máximas loop.Amount) pidiendo los ganadores hasta que o bien reciban los ganadores, o se rindan. Se decidió realizarlo de esta manera ya que en un servidor secuencial me resultaba extraño posponer la ejecución de una de las agencias y dejarlo en espera.

A su vez, durante todo el ejercicio se mandan los distintos batches en distintas conexiones. Esto permite que la agencia pueda crearse múltiples conexiones y enviarlas en nombre de su número de agencia. Permite además que el servidor vaya tomando los pedidos de las agencias aleatoriamente y que no tenga que terminar con uno para atender a la otra.

### Cómo ejecutarlo
Como prerrequisitos, se pide que si no se tiene un archivo compose con 5 clientes, se puede crear uno ejecutando el comando:

```bash
./generar-compose.sh docker-compose-dev.yaml 5
```
Además, se debe asegurar de contar con una carpeta de data con las siguientes rutas:
```bash
/.data/agency-<id cliente n>.csv
```
con 1 <= n <= 5, teniendo la carpeta de data en la ruta al repositorio, y siendo el id del cliente el que se puede hallar en el compose como environment: CLI_ID.
Para ejecutarlo luego de estos requisitos se debe seguir los mismos pasos que en el ejercicio 5. 

# Parte 3
## Ejercicio 8
Para este último ejercicio, en donde se requería la conversión de nuestro servidor secuencial a una solución en donde se puedan tomar multiples clientes, se desarrolló una solución sencilla haciendo uso de una thread pool con 5 hilos, tal como la cantidad de agencias. La idea es que cada hilo se mappee a un cliente en particular y lleve el rastro de este hasta su finalización. La razón por la que se hace uso de una thread pool y no se lanzan hilos ante la llegadas de cliente es para hacer el servidor más estable, y que eventualmente si se desea tener más clientes no se deba lanzar un hilo por cada uno, y además para dificultar la posibilidad de un ataque de DoS.

Para ejecutar tareas en la ThreadPool existen dos métodos:
* apply
* apply_async
Se decidió trabajar con el segundo ya que el primero espera a que la función retorne para correr otra tarea, lo cual para este problema no sirve, pues nos otorga en definitiva procesamiento secuencial.

Para poder sincronizar estos procesos se hace uso principalmente de dos herramientas.
#### Una variable de condición 
Esta se usa para llevar registro de las agencias que han finalizado. Al recibir un mensaje de finalización de apuestas, el servidor almacena en el set de agencias a esta. Se eligió un set y no un contador para que de igual cuántas veces la misma agencia diga que ha finalizado, se tomará solo la primera vez. Una vez recibido este mensaje, si la agencia no estaba siendo trackeada como parte de las finalizadas, entonces habiendo tomado esta condvar, se la agrega al set y revisa si llegó a 5. Si llegó a 5, entonces "despierta" (notify all) a todos los procesos que se encontraban a la espera de que las otras agencias cierren. Caso contrario, simplemente sigue con su procesamiento.

Cuando una agencia pide al servidor los ganadores que apostaron en su agencia, el hilo de este toma la variable de condición (que tiene un lock asociado internamente) y revisa si las 5 agencias ya mandaron los resultados. En caso de que esto no sea así, entonces se "duerme" haciendo un wait, hasta que se realice el notify all por parte de otro hilo.   

#### Lock
Este se usa para proteger el archivo de apuestas, tanto para escritura como para lectura. Esto implica que se debe tomar el lock antes de llamar a las funciones `store_bets` y `load_bets` (get_winners). Este lock a su vez protege a la variable asociada al archivo: `winners`. La idea es que, en lugar de que cada uno de los hilos cargue el archivo en memoria, procese a los ganadores y tenga su propia variable local con estos, que busque cada vez que un cliente se lo pide, que ya sea parte de lo que es el servidor, y que el primero que llegue tras el notify_all sea quien va a cargar todos los registros a memoria. Cuando el resto quiera buscar los ganadores, si ya esta variable está setteada entonces simplemente la usará para buscar las agencias ganadoras y guardarselas en una variable temporal. Es importante destacar que la variable winners no se debe usar fuera de la sección crítica, por lo que se la copia para guardar los ganadores de cada agencia y no se la use directamente, ni aunque sea para solo lectura.   

### Como ejecutarlo
Este ejercicio no requiere más que los anteriores. Se debe asegurarse de tener 5 clientes para que no se queden infinitamente esperando una notificación que no llegara, tal como en el ejercicio 7.
 
# Referencias
[1] 
[Instalar netcat en contenedor](https://stackoverflow.com/questions/52570028/i-am-unable-to-execute-netcat-command-within-docker-bash-terminal)

[2]
[Solucionar problema: 'netcat' has no installation candidate](https://forums.docker.com/t/package-netcat-has-no-installation-candidate-how-to-fix-this/136541)

[3] 
[Correr comandos de terminal en docker container](https://www.warp.dev/terminus/docker-run-bash)

[4]
[Librería para el manejo de señales en go](https://go.dev/pkg/os/signal/?m=old)

[5] 
[Ejemplos de manejo de señales en go](https://gobyexample.com/signals)

[6] 
[Ejemplos de waitgroups en go](https://gobyexample.com/waitgroups)

[7]
[Modificar offset de lectura en un archivo](https://pkg.go.dev/os#File.Seek)

[8]
[Mover elementos de un slice al otro](https://stackoverflow.com/questions/74938295/how-to-move-elements-from-one-slice-to-another)

[9]
[Slicing en python](https://stackoverflow.com/questions/9490058/why-does-substring-slicing-with-index-out-of-range-work)

[10]
[Multiprocessing library python](https://docs.python.org/3/library/multiprocessing.html)