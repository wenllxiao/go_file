```mysql
#查看每张表大小G显示
SELECT TABLE_SCHEMA, TABLE_NAME, DATA_LENGTH/(1024*1024*1024) AS TABLE_SIZE_GB
FROM information_schema.TABLES order by DATA_LENGTH desc
```

```mysql
#创建分区表 参考文档：https://www.jianshu.com/p/19aae55983f0
DROP PROCEDURE IF EXISTS auto_create_process_test_item_partition;
DELIMITER //
CREATE PROCEDURE auto_create_process_test_item_partition(IN manu_spc VARCHAR(64), IN table_name VARCHAR(64))
BEGIN
    DECLARE done INT DEFAULT FALSE;
    DECLARE cur_date DATE;
    DECLARE PARTITIONNAME VARCHAR(9);
    DECLARE PARTITIONRANGE INT UNSIGNED;
    DECLARE ROWS_CNT INT;

    DECLARE cur CURSOR FOR
        SELECT CURDATE() + INTERVAL n DAY AS new_date
        FROM (SELECT 0 n UNION ALL SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3 UNION ALL SELECT 4) num_seq;

    DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = TRUE;
    DECLARE EXIT HANDLER FOR SQLEXCEPTION
        BEGIN
            -- 如果发生错误，则关闭游标
            CLOSE cur;
            -- 可以添加其他的错误处理逻辑
            RESIGNAL; -- 重新抛出异常
        END;

    OPEN cur;

    read_loop: LOOP
        FETCH cur INTO cur_date;
        IF done THEN
            LEAVE read_loop;
        END IF;

        SET PARTITIONNAME = DATE_FORMAT(cur_date, 'p%Y%m%d');
        SET PARTITIONRANGE = UNIX_TIMESTAMP(cur_date + INTERVAL 1 DAY);
        -- 查询指定分区是否存在
        SELECT COUNT(*)
        INTO ROWS_CNT
        FROM information_schema.partitions
        WHERE table_schema = manu_spc
          AND table_name = table_name
          AND partition_name = PARTITIONNAME;
        -- 判断分区是否存在，如果不存在则创建
        IF ROWS_CNT = 0 THEN
            SET @SQL = CONCAT('ALTER TABLE `', manu_spc, '`.`', table_name, '` ADD PARTITION (PARTITION ', PARTITIONNAME,
                              ' VALUES LESS THAN (', PARTITIONRANGE, ') ENGINE = InnoDB);');

            PREPARE STMT FROM @SQL;
            EXECUTE STMT;
            DEALLOCATE PREPARE STMT;
            SELECT CONCAT('Partition `', PARTITIONNAME, '` for table `', manu_spc, '.', table_name, '` created successfully') AS result;
        ELSE
            SELECT CONCAT('Partition `', PARTITIONNAME, '` for table `', manu_spc, '.', table_name, '` already exists') AS result;
        END IF;
    END LOOP;
    CLOSE cur;
END //
DELIMITER ;


drop event if exists event_create_process_test_item_partition;
create event event_create_process_test_item_partition on schedule
    every '1' DAY
        #从现在开始
        STARTS now()
    on completion preserve
    enable
    do
    #调用分区存储过程
    call auto_create_process_test_item_partition('manu_spc', 'process_test_item');

call auto_create_process_test_item_partition('manu_spc', 'process_test_item');


SHOW PROCESSLIST;
SHOW EVENTS;
```


