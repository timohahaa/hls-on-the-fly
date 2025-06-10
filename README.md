# Генерация hls-манифестов и шифрование видео "налету"
Proof-of-concept создания hls-манифестов налету из fMP4 файлов + стриминг шифрованных видео налету

Для манифестов byte-ranges и длительность фрагментов вычисляется парсингом fMP4 файла.

Видео шифруются налету из нешифрованных fMP4 файлов


Note: шифрованные видео не будут работать в `hls.js` :_), но будут в нативном hls

Examples:
- http://localhost:8001/video-1/master.m3u8
- http://localhost:8001/video-2/master.m3u8

Шифрованные:
- http://localhost:8001/video-1/master.m3u8?ecnrypt=true
- http://localhost:8001/video-2/master.m3u8?encrypt=true
